package service

import (
	"Ralf/domen"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetNewTask читает файл задач и возвращает первую задачу со статусом new.
// Если задача не найдена или произошла ошибка ввода-вывода, возвращается соответствующая ошибка.
func GetNewTask(path string) (domen.Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return domen.Task{}, fmt.Errorf("не удалось открыть файл задач: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var tasks []domen.Task
	var currentTask map[string]string
	inTask := false

	// Построчный разбор файла
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Обнаружено начало новой задачи
		if strings.HasPrefix(line, "начало задачи:") {
			inTask = true
			currentTask = make(map[string]string)
			continue
		}

		// Обнаружен конец задачи
		if strings.HasPrefix(line, "конец задачи.") && inTask {
			inTask = false
			task, err := parseTaskFromMap(currentTask)
			if err != nil {
				// При ошибке парсинга одной задачи прерываем выполнение,
				// так как файл может быть повреждён.
				return domen.Task{}, fmt.Errorf("ошибка парсинга задачи: %w", err)
			}
			tasks = append(tasks, task)
			continue
		}

		// Если мы внутри задачи, обрабатываем строку с ключом и значением
		if inTask {
			// Разделяем по первому двоеточию, чтобы отделить ключ от значения
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				// Строка не содержит двоеточия — игнорируем (возможно, часть предыдущего значения)
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			currentTask[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return domen.Task{}, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	// Поиск первой задачи со статусом new
	for _, task := range tasks {
		if task.Status == domen.StatusNew {
			return task, nil
		}
	}

	return domen.Task{}, errors.New("не найдено задач со статусом new")
}

// parseTaskFromMap преобразует набор пар «ключ-значение» в структуру Task.
// Ключи соответствуют русскоязычным заголовкам из файла.
func parseTaskFromMap(data map[string]string) (domen.Task, error) {
	var task domen.Task
	//var err error

	for key, value := range data {
		switch key {
		case "номер задачи":
			num, convErr := strconv.Atoi(value)
			if convErr != nil {
				return domen.Task{}, fmt.Errorf("неверный формат номера задачи: %w", convErr)
			}
			task.Num = num
		case "описание задачи":
			task.Description = value
		case "важные моменты":
			task.ImportantInfo = value
		case "ожидаемый результат":
			task.ExpectResult = value
		case "тестовые данные":
			task.TestsValue = value
		case "сигнатура функции":
			task.FuncSignature = value
		case "статус выполнения":
			task.Status = domen.TaskStatus(value)
		default:
			// Неизвестные ключи игнорируются, что позволяет расширять формат без поломки парсера
		}
	}

	// Проверка наличия обязательных полей не производится,
	// так как в задании не указаны обязательные требования.
	// При необходимости можно добавить проверку, например, на Num==0 и возвращать ошибку.
	return task, nil
}
