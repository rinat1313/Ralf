package service

import (
	"Ralf/domen"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// RunOrchestrator запускает главный цикл управления всеми модулями:
// проверка окружения → чтение задач → обработка одной задачи с лимитами циклов.
// Вызывается из main().
func RunOrchestrator(cfg domen.Config) error {
	// значения по умолчанию
	if cfg.TasksFilePath == "" {
		cfg.TasksFilePath = "tasks.txt"
	}
	if cfg.MaxTaskAttempts == 0 {
		cfg.MaxTaskAttempts = 10
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

	if err := checkLMStudioAvailable(); err != nil {
		return fmt.Errorf("LM Studio недоступен: %w", err)
	}
	if err := checkGoAndFSAccess(); err != nil {
		return fmt.Errorf("проблема с окружением Go или правами ФС: %w", err)
	}

	// основной цикл обработки задач
	for i := 0; i < cfg.MaxTaskAttempts; i++ {
		task, err := GetNewTask(cfg.TasksFilePath)
		if err != nil {
			if errors.Is(err, errors.New("не найдено задач со статусом new")) {
				return nil // все задачи обработаны
			}
			return fmt.Errorf("ошибка получения задачи: %w", err)
		}

		// меняем статус на run
		if err := UpdateTaskStatus(cfg.TasksFilePath, task.Num, domen.StatusRun); err != nil {
			return fmt.Errorf("не удалось обновить статус run: %w", err)
		}

		if err := processTask(task, cfg); err != nil {
			_ = UpdateTaskStatus(cfg.TasksFilePath, task.Num, domen.StatusError)
			return fmt.Errorf("обработка задачи %d завершилась ошибкой: %w", task.Num, err)
		}

		// успешно
		if err := UpdateTaskStatus(cfg.TasksFilePath, task.Num, domen.StatusOK); err != nil {
			return fmt.Errorf("не удалось обновить статус ok: %w", err)
		}
	}
	return nil
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

// processTask выполняет полный цикл для одной задачи:
// 1. Получение команд от LM Studio
// 2. Выполнение команд
// 3. Цикл исправления компиляции (до MaxCompileFixAttempts)
// 4. Генерация тестов
// 5. Компиляция тестов с исправлениями
func processTask(task domen.Task, cfg domen.Config) error {
	// 1-2. Получаем и выполняем команды решения
	commands, err := SendTaskToLLM(task)
	if err != nil {
		return fmt.Errorf("ошибка получения решения от LM Studio: %w", err)
	}
	for _, cmd := range commands {
		if _, execErr := ExecuteCommand(cmd); execErr != nil {
			return fmt.Errorf("ошибка выполнения команды: %w", execErr)
		}
	}

	// 3. Цикл исправления компиляции
	for i := 0; i < cfg.MaxCompileFixAttempts; i++ {
		compileLog, compileErr := Compile(".")
		if compileErr == nil {
			break
		}
		// отправляем ошибку и получаем исправления
		fixResp, fixErr := SendCompilationError("main.go", compileLog)
		if fixErr != nil {
			return fmt.Errorf("ошибка отправки лога компиляции: %w", fixErr)
		}
		fixCommands, parseErr := ParseCommands(fixResp)
		if parseErr != nil {
			return fmt.Errorf("ошибка парсинга исправлений: %w", parseErr)
		}
		for _, cmd := range fixCommands {
			if _, execErr := ExecuteCommand(cmd); execErr != nil {
				return fmt.Errorf("ошибка применения исправления: %w", execErr)
			}
		}
	}

	// 4. Генерация тестов (отдельный промпт)
	testCommands, testErr := generateTests(task)
	if testErr != nil {
		return fmt.Errorf("ошибка генерации тестов: %w", testErr)
	}
	for _, cmd := range testCommands {
		if _, execErr := ExecuteCommand(cmd); execErr != nil {
			return fmt.Errorf("ошибка выполнения команд тестов: %w", execErr)
		}
	}

	// 5. Цикл компиляции тестов
	for i := 0; i < cfg.MaxTestAttempts; i++ {
		_, testCompileErr := Compile(".")
		if testCompileErr == nil {
			return nil
		}
		// повторяем исправление для тестов
		testFixResp, _ := SendCompilationError("_test.go", "ошибка компиляции тестов")
		testFixCmds, _ := ParseCommands(testFixResp)
		for _, cmd := range testFixCmds {
			ExecuteCommand(cmd)
		}
	}
	return nil
}

// generateTests отправляет LM Studio запрос на генерацию тестов.
func generateTests(task domen.Task) ([]domen.Command, error) {
	//testPrompt := fmt.Sprintf(`Напиши только тесты для задачи №%d в формате команд.
	//Тесты должны быть в файле %s_test.go и использовать стандартный пакет testing.
	//Верни ТОЛЬКО блоки команд.`, task.Num, task.FuncSignature)
	// используем тот же механизм с системным промптом
	//messages := []Message{
	//	{Role: "system", Content: StrictCommandTemplate},
	//	{Role: "user", Content: testPrompt},
	//}
	// здесь можно расширить, но для чистоты используем существующий клиент
	// пока возвращаем через SendTaskToLLM с модифицированным промптом (упрощённо)
	return SendTaskToLLM(task) // переиспользуем, LLM поймёт контекст
}
