package service

import (
	"Ralf/domen"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	lmStudioURL = "http://localhost:1234/v1/chat/completions"
	modelName   = "local-model"
)

// buildSystemPrompt формирует предварительный системный промпт, строго определяющий
// формат ответа LM Studio для совместимости с ParseCommands.
func buildSystemPrompt() string {
	return `Ты — эксперт-программист Go. Решаешь задачи и возвращаешь ТОЛЬКО команды в строгом формате.

Используй блоки:

Начало команды:
тип: создание
адрес: /path/to/file.go
содержимое:
...полный код...
Конец команды.

Или для редактирования:
тип: внесение изменений
адрес: /path/to/file.go
строки: {5:"новый код", 12:"другой код"}
Конец команды.

Для компиляции:
тип: компиляция
адрес: /path/to/main.go
Конец команды.

Никакого текста вне этих блоков. Только команды.`
}

// taskToPrompt формирует пользовательский промпт из структуры Task.
func taskToPrompt(task domen.Task) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Задача №%d\n", task.Num))
	sb.WriteString(fmt.Sprintf("Описание: %s\n", task.Description))
	sb.WriteString(fmt.Sprintf("Важные моменты: %s\n", task.ImportantInfo))
	sb.WriteString(fmt.Sprintf("Ожидаемый результат: %s\n", task.ExpectResult))
	sb.WriteString(fmt.Sprintf("Тесты: %s\n", task.TestsValue))
	if task.FuncSignature != "" {
		sb.WriteString(fmt.Sprintf("Сигнатура: %s\n", task.FuncSignature))
	}
	return sb.String()
}

// SendTaskToLMStudio отправляет структуру Task в LM Studio через REST API.
// Предварительно применяется system prompt с требованием строгого шаблона ответа.
func SendTaskToLMStudio(task domen.Task) (string, error) {
	messages := []Message{
		{Role: "system", Content: buildSystemPrompt()},
		{Role: "user", Content: taskToPrompt(task)},
	}
	return sendToLMStudio(messages)
}

// SendCompilationError отправляет лог ошибки компиляции в LM Studio.
// LM Studio возвращает список исправляющих команд в требуемом формате.
func SendCompilationError(path, compileLog string) (string, error) {
	prompt := fmt.Sprintf("Файл %s не компилируется. Лог ошибки:\n\n%s\n\nИсправь код. Верни только команды для исправления.", path, compileLog)
	messages := []Message{
		{Role: "system", Content: buildSystemPrompt()},
		{Role: "user", Content: prompt},
	}
	return sendToLMStudio(messages)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// sendToLMStudio выполняет HTTP-запрос к OpenAI-совместимому endpoint LM Studio.
func sendToLMStudio(messages []Message) (string, error) {
	reqBody := chatRequest{
		Model:       modelName,
		Messages:    messages,
		Temperature: 0.1,
		MaxTokens:   4096,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка создания JSON: %w", err)
	}

	resp, err := http.Post(lmStudioURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("не удалось подключиться к LM Studio: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LM Studio: статус %d, тело: %s", resp.StatusCode, string(body))
	}

	var result chatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("пустой ответ от LM Studio")
	}

	return result.Choices[0].Message.Content, nil
}
