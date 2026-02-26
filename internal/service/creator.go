package service

import (
	"Ralf/domen"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UpdateTaskStatus изменяет статус задачи с указанным номером в файле.
// Возвращает nil при успешной замене, иначе ошибку с описанием.
func UpdateTaskStatus(filePath string, taskNum int, newStatus domen.TaskStatus) error {
	// 1. Открываем исходный файл для чтения
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл для чтения: %w", err)
	}
	defer inputFile.Close()

	// 2. Создаём временный файл для записи обновлённого содержимого
	tempFile, err := os.CreateTemp("", "tasks_*.txt")
	if err != nil {
		return fmt.Errorf("не удалось создать временный файл: %w", err)
	}
	tempFileName := tempFile.Name()
	defer tempFile.Close()

	scanner := bufio.NewScanner(inputFile)
	var inTask bool
	var currentTaskNum int
	var taskLines []string
	taskFound := false
	statusUpdated := false

	// 3. Построчно читаем исходный файл и пишем во временный
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Обнаружено начало задачи
		if strings.HasPrefix(trimmedLine, "начало задачи:") {
			inTask = true
			taskLines = []string{line} // начинаем собирать строки задачи
			continue
		}

		// Если мы внутри задачи, добавляем строку в буфер
		if inTask {
			taskLines = append(taskLines, line)

			// Пытаемся извлечь номер задачи из текущей строки
			if strings.Contains(trimmedLine, "номер задачи:") {
				parts := strings.SplitN(trimmedLine, ":", 2)
				if len(parts) == 2 {
					numStr := strings.TrimSpace(parts[1])
					if num, err := strconv.Atoi(numStr); err == nil {
						currentTaskNum = num
					}
				}
			}

			// Проверяем конец задачи
			if strings.HasPrefix(trimmedLine, "конец задачи.") {
				// Задача завершена, проверяем, нужно ли обновить статус
				if currentTaskNum == taskNum {
					taskFound = true
					// Обновляем строку статуса в собранных строках задачи
					updatedLines := make([]string, 0, len(taskLines))
					for _, taskLine := range taskLines {
						if strings.Contains(taskLine, "статус выполнения:") {
							// Заменяем старый статус на новый
							newLine := updateStatusInLine(taskLine, newStatus)
							updatedLines = append(updatedLines, newLine)
							statusUpdated = true
						} else {
							updatedLines = append(updatedLines, taskLine)
						}
					}
					// Записываем обновлённые строки задачи во временный файл
					for _, l := range updatedLines {
						_, err := tempFile.WriteString(l + "\n")
						if err != nil {
							return fmt.Errorf("ошибка записи во временный файл: %w", err)
						}
					}
				} else {
					// Это не наша задача — записываем как есть
					for _, l := range taskLines {
						_, err := tempFile.WriteString(l + "\n")
						if err != nil {
							return fmt.Errorf("ошибка записи во временный файл: %w", err)
						}
					}
				}
				// Сбрасываем состояние для следующей задачи
				inTask = false
				currentTaskNum = 0
				taskLines = nil
				continue
			}
		} else {
			// Строка вне задачи — просто копируем
			_, err := tempFile.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("ошибка записи во временный файл: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения исходного файла: %w", err)
	}

	// 4. Проверяем, что задача была найдена и статус обновлён
	if !taskFound {
		return fmt.Errorf("не получилось поменять статус задачи № %d в файле %s: задача не найдена", taskNum, filePath)
	}
	if !statusUpdated {
		return fmt.Errorf("не получилось поменять статус задачи № %d в файле %s: поле статуса не найдено", taskNum, filePath)
	}

	// 5. Закрываем файлы перед заменой
	inputFile.Close()
	tempFile.Close()

	// 6. Заменяем исходный файл временным
	if err := os.Rename(tempFileName, filePath); err != nil {
		return fmt.Errorf("не удалось заменить исходный файл: %w", err)
	}

	return nil
}

// updateStatusInLine заменяет старое значение статуса на новое в строке.
func updateStatusInLine(line string, newStatus domen.TaskStatus) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return line // неожиданный формат, оставляем как есть
	}
	// Оставляем ключ без изменений, подставляем новое значение статуса
	return fmt.Sprintf("%s:%s", parts[0], string(newStatus))
}
