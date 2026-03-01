package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"Ralf/domen"
)

const StrictCommandTemplate = `Ты — эксперт-программист Go.

Твоя ЕДИНСТВЕННАЯ задача — выполнять задачу и возвращать ТОЛЬКО валидный JSON-массив объектов.

ОБЯЗАТЕЛЬНО: Используй ТОЧНО такие значения в поле "Type" — без сокращений, без синонимов:
- "создание" (не "create", не "создать")
- "удаление" (не "delete", не "удалить")
- "внесение изменений" (не "изменение", не "edit", не "изменить")
- "добавление строк" (не "add lines")
- "удаление строк" (не "delete lines")
- "копирование" (не "copy")
- "перемещение" (не "move")
- "чтение" (не "read")
- "компиляция" (не "compile")

Пример правильного поля Type:
"Type": "внесение изменений"   ← именно так, полностью

ВАЖНЕЙШЕЕ ТРЕБОВАНИЕ №1: ВСЕ пути начинаются с tasks/task_<номер задачи>/
Примеры: tasks/task_1/main.go, tasks/task_1/greeting_test.go

ПРАВИЛО №1: Ответ — ТОЛЬКО JSON-массив. Никакого текста, markdown, ±json.
ПРАВИЛО №2: Пример ответа:

[
{
"Type": "создание",
"Path": "tasks/task_1/main.go",
"Content": "package main\\n\\nimport \\"fmt\\"\\n\\nfunc Greeting(name string) string {\\n\\tif name == \\"\\" {\\n\\t\\treturn \\"Hello, World!\\"\\n\\t}\\n\\treturn \\"Hello, \\" + name + \\"!\\"\\n}"
}
]

Выполни задачу и верни ТОЛЬКО JSON-массив.`

// LLMClient управляет взаимодействием с LM Studio через OpenAI-compatible API.
type LLMClient struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

func NewLLMClient() *LLMClient {
	return &LLMClient{
		BaseURL: "http://localhost:1234/v1",
		Model:   "local-model",
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// SendTaskToLLM отправляет структуру Task в LM Studio и возвращает список parsed команд.
func SendTaskToLLM(task domen.Task) ([]domen.Command, error) {
	client := NewLLMClient()
	return client.sendTask(task)
}

func (c *LLMClient) sendTask(task domen.Task) ([]domen.Command, error) {
	userPrompt := fmt.Sprintf(`Задача №%d

Описание задачи: %s
Важные моменты: %s
Ожидаемый результат: %s
Тестовые данные: %s
Сигнатура функции: %s

Реализуй задачу строго по шаблону выше.
Все файлы должны находиться внутри папки task_%d.
Если нужно создать тесты — используй task_%d/<имя_функции>_test.go`,
		task.Num,
		task.Description,
		task.ImportantInfo,
		task.ExpectResult,
		task.TestsValue,
		task.FuncSignature,
		task.Num,
		task.Num)

	reqBody := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": StrictCommandTemplate},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.0,
		"top_p":       1.0,
		"max_tokens":  16384,
		"stream":      false,
	}
	fmt.Println("Начинаем процесс маршалирование запроса от llm.")
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("не удалось маршалировать запрос: %w", err)
	}
	fmt.Println("Начинаем процесс соединения с LM Studio.")
	resp, err := c.HTTPClient.Post(c.BaseURL+"/chat/completions", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка соединения с LM Studio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LM Studio вернул код %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	fmt.Println("Начинаем парсинг JSON ответа.")
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON ответа: %w", err)
	}

	if len(apiResp.Choices) == 0 || apiResp.Choices[0].Message.Content == "" {
		return nil, errors.New("LM Studio вернул пустой ответ")
	}
	fmt.Println("Начинаем процесс парсинга команд:")
	llmOutput := apiResp.Choices[0].Message.Content
	return ParseCommands(llmOutput)
}
