package service

import (
	"Ralf/domen"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunOrchestrator запускает обработку ВСЕХ задач со статусом "new"
func RunOrchestrator(cfg domen.Config) error {
	// значения по умолчанию
	if cfg.TasksFilePath == "" {
		cfg.TasksFilePath = "tasks.txt"
	}
	if cfg.MaxCompileFixAttempts == 0 {
		cfg.MaxCompileFixAttempts = 10
	}
	if cfg.MaxTestAttempts == 0 {
		cfg.MaxTestAttempts = 10
	}
	if cfg.WorkingDir == "" {
		cfg.WorkingDir = "."
	}

	fmt.Println("Приступаем к этапу анализа доступов.")
	if err := checkLMStudioAvailable(); err != nil {
		return fmt.Errorf("LM Studio недоступен: %w", err)
	}
	if err := checkGoAndFSAccess(); err != nil {
		return fmt.Errorf("проблема с окружением Go или правами ФС: %w", err)
	}

	fmt.Println("Начинаем цикл обработки задач.")

	processed := 0
	for {
		task, err := GetNewTask(cfg.TasksFilePath)
		if err != nil {
			if errors.Is(err, errors.New("не найдено задач со статусом new")) {
				fmt.Printf("Все задачи обработаны успешно. Обработано задач: %d\n", processed)
				return nil
			}
			return fmt.Errorf("ошибка получения задачи: %w", err)
		}

		fmt.Printf("Получили задачу в тексте %d\n", task.Num)
		fmt.Println("Меняем статус на run.")

		if err := UpdateTaskStatus(cfg.TasksFilePath, task.Num, domen.StatusRun); err != nil {
			return fmt.Errorf("не удалось обновить статус run: %w", err)
		}

		fmt.Println("Приступаем к Process task:")
		if err := processTask(task, cfg); err != nil {
			_ = UpdateTaskStatus(cfg.TasksFilePath, task.Num, domen.StatusError)
			fmt.Printf("Задача %d завершилась ошибкой: %v\n", task.Num, err)
			// Продолжаем обработку следующих задач, не выходим!
			continue
		}

		fmt.Println("Меняем статус на ok.")
		if err := UpdateTaskStatus(cfg.TasksFilePath, task.Num, domen.StatusOK); err != nil {
			return fmt.Errorf("не удалось обновить статус ok: %w", err)
		}

		processed++
	}
}

// checkLMStudioAvailable проверяет доступность LM Studio простым запросом.
func checkLMStudioAvailable() error {
	_, err := http.Get("http://localhost:1234/v1/models")
	if err != nil {
		return err
	}
	return nil
}

// checkGoAndFSAccess проверяет наличие go и права на запись/исполнение.
func checkGoAndFSAccess() error {
	if _, err := exec.LookPath("go"); err != nil {
		return err
	}
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return err
	}
	return os.RemoveAll("tmp")
}

// processTask выполняет полный цикл для одной задачи
func processTask(task domen.Task, cfg domen.Config) error {
	fmt.Println("Отправляем структуру task в llm.")

	// 1. Основной код + тесты
	commands, err := SendTaskToLLM(task)
	if err != nil {
		return fmt.Errorf("ошибка получения решения от LM Studio: %w", err)
	}
	fmt.Println("Начинаем выполнять полученные команды:")
	for _, cmd := range commands {
		if _, execErr := ExecuteCommand(cmd); execErr != nil {
			return fmt.Errorf("ошибка выполнения команды: %w", execErr)
		}
	}

	// 2. Цикл исправления компиляции (с номером попытки)
	for i := 0; i < cfg.MaxCompileFixAttempts; i++ {
		compileLog, compileErr := Compile(".")
		if compileErr == nil {
			break
		}

		fmt.Printf("Попытка исправления %d/%d...\n", i+1, cfg.MaxCompileFixAttempts)

		fixResp, fixErr := SendCompilationError(
			"prog/main.go",
			compileLog,
			i+1, // ← передаём номер попытки
		)
		if fixErr != nil {
			fmt.Println("Не удалось отправить ошибку компиляции")
			break
		}

		fixCommands, parseErr := ParseCommands(fixResp)
		if parseErr != nil {
			fmt.Println("Не удалось распарсить исправления")
			break
		}

		for _, cmd := range fixCommands {
			ExecuteCommand(cmd)
		}
	}

	// 3. Генерация тестов
	testCommands, testErr := generateTests(task)
	if testErr != nil {
		return fmt.Errorf("ошибка генерации тестов: %w", testErr)
	}
	for _, cmd := range testCommands {
		if _, execErr := ExecuteCommand(cmd); execErr != nil {
			return fmt.Errorf("ошибка выполнения команд тестов: %w", execErr)
		}
	}

	// 4. Компиляция тестов
	for i := 0; i < cfg.MaxTestAttempts; i++ {
		_, testCompileErr := Compile(".")
		if testCompileErr == nil {
			return nil // всё успешно
		}
		// исправление тестов (можно тоже через SendCompilationError, но пока оставляем как было)
		testFixResp, _ := SendCompilationError("_test.go", "ошибка компиляции тестов", 1)
		testFixCmds, _ := ParseCommands(testFixResp)
		for _, cmd := range testFixCmds {
			ExecuteCommand(cmd)
		}
	}
	return nil
}

// generateTests отправляет LM Studio запрос на генерацию ТОЛЬКО тестов
func generateTests(task domen.Task) ([]domen.Command, error) {
	testPrompt := fmt.Sprintf(`Это уже решённая задача №%d.
Сигнатура функции: %s

Теперь напиши ТОЛЬКО тесты в формате JSON-массива команд.
Тесты должны быть в файле prog/%s_test.go
Используй пакет testing и таблицу тестов.
Не трогай основной код — только создавай/редактируй тестовый файл.

Пример:
[
  {
    "Type": "создание",
    "Path": "prog/%s_test.go",
    "Content": "package main_test\\n\\nimport (\\n\\t\\"testing\\"\\n)\\n\\nfunc TestGreeting(t *testing.T) {\\n\\t// тесты здесь\\n}"
  }
]

Верни ТОЛЬКО JSON-массив.`,
		task.Num,
		task.FuncSignature,
		strings.TrimSuffix(filepath.Base(task.FuncSignature), " string) string"), // имя функции без сигнатуры
		strings.TrimSuffix(filepath.Base(task.FuncSignature), " string) string"))

	// Создаём временный task только для тестов (чтобы не менять оригинальный)
	testTask := task
	testTask.Description = testPrompt // переопределяем описание → LLM поймёт, что нужно тесты

	return SendTaskToLLM(testTask)
}
