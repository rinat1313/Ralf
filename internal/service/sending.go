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

const StrictCommandTemplate = `Ты — эксперт-программист Go с 5-летним опытом.

ВАЖНЕЙШЕЕ ТРЕБОВАНИЕ №1: ВСЕ пути к файлам ДОЛЖНЫ начинаться с tasks/task_<номер задачи>/
Примеры правильных путей:
- tasks/task_1/main.go
- tasks/task_1/greeting_test.go
- tasks/task_2/add.go
- tasks/task_3/is_even_test.go

Никогда не используй пути без tasks/task_N/ в начале — это критическая ошибка!

Твоя ЕДИНСТВЕННАЯ задача — выполнять полученную задачу и вернуть ТОЛЬКО валидный JSON-массив.
Важно: Все файлы — внутри tasks/task_<N>/, где N — номер текущей задачи.
ПРАВИЛО №1: Ответ — строго JSON-массив объектов. Никакого текста до или после.
ПРАВИЛО №2: Запрещено: markdown, json, объяснения, любые комментарии.
ПРАВИЛО №3: Пример ответа (все пути начинаются с tasks/task_N/):

[
{
"Type": "создание",
"Path": "tasks/task_1/main.go",
"Content": "package main\\n\\nimport \\"fmt\\"\\n\\nfunc Greeting(name string) string {\\n\\tif name == \\"\\" {\\n\\t\\treturn \\"Hello, World!\\"\\n\\t}\\n\\treturn \\"Hello, \\" + name + \\"!\\"\\n}"
},
{
"Type": "создание",
"Path": "tasks/task_1/greeting_test.go",
"Content": "package main_test\\n\\nimport (\\"testing\\")\\n\\nfunc TestGreeting(t *testing.T) {\\n\\t// тесты здесь\\n}"
}
]

ПРАВИЛО №4: Используй только поля: "Type", "Path", "Content", "Lines", "SrcPath", "DstPath".
ПРАВИЛО №5: "Lines" — объект с ключами-строками (номера строк).

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

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("не удалось маршалировать запрос: %w", err)
	}

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

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON ответа: %w", err)
	}

	if len(apiResp.Choices) == 0 || apiResp.Choices[0].Message.Content == "" {
		return nil, errors.New("LM Studio вернул пустой ответ")
	}

	llmOutput := apiResp.Choices[0].Message.Content
	return ParseCommands(llmOutput)
}
