package service

import (
	"Ralf/domen"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	lmStudioURL = "http://localhost:1234/v1/chat/completions"
	modelName   = "local-model"
)

// buildSystemPrompt формирует предварительный системный промпт, строго определяющий
// формат ответа LM Studio для совместимости с ParseCommands.
func buildSystemPrompt() string {
	return `Ты — эксперт-программист Go.
Твоя ЕДИНСТВЕННАЯ задача — выполнять задачу и возвращать ТОЛЬКО валидный JSON-массив объектов.

ВАЖНЕЙШЕЕ ПРАВИЛО: ВСЕ пути начинаются с tasks/task_<номер задачи>/
Примеры: tasks/task_1/main.go, tasks/task_1/greeting_test.go

Никогда не используй абсолютные пути или пути без tasks/task_N/

ПРАВИЛО №1: Ответ — ТОЛЬКО JSON-массив. Никакого текста, markdown, json.
		ПРАВИЛО №2: Каждый объект — одна команда. Пример:

[
{
"Type": "создание",
"Path": "tasks/task_1/main.go",
"Content": "package main\\n\\nimport \\"fmt\\"\\n\\nfunc Greeting(name string) string {\\n\\tif name == \\"\\" {\\n\\t\\treturn \\"Hello, World!\\"\\n\\t}\\n\\treturn \\"Hello, \\" + name + \\"!\\"\\n}"
}
]

ПРАВИЛО №3: Используй поля: "Type", "Path", "Content", "Lines", "SrcPath", "DstPath".
ПРАВИЛО №4: "Lines" — объект с ключами-строками (номера строк как строки).
ПРАВИЛО №5: В Content используй \\n для переносов строк.

Выполни задачу и верни ТОЛЬКО JSON-массив.`
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

// SendCompilationError — теперь LLM видит текущий код файла и понимает, что это повторная попытка
func SendCompilationError(path, compileLog string, attempt int) (string, error) {
	// Читаем текущий код файла
	currentCode := ""
	if data, err := os.ReadFile(path); err == nil {
		currentCode = string(data)
	}

	prompt := fmt.Sprintf(`Это ПОПЫТКА ИСПРАВЛЕНИЯ №%d (максимум 10).

Файл: %s

ТЕКУЩИЙ КОД (который сейчас НЕ компилируется):
go
	%s
	Лог ошибки компиляции:
	%s
	Предыдущие исправления НЕ СРАБОТАЛИ.
		НЕ повторяй предыдущий код!
		Внеси реальные изменения, чтобы файл скомпилировался без ошибок.
		Верни ТОЛЬКО JSON-массив команд (как всегда).`,
		attempt, path, currentCode, compileLog)
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
