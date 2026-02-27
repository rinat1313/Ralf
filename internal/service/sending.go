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
Твоя ЕДИНСТВЕННАЯ задача — выполнять полученную задачу ИСКЛЮЧИТЕЛЬНО через блоки команд.

ПРАВИЛО №1: Ты НЕ ИМЕЕШЬ ПРАВА выводить НИЧЕГО кроме блоков команд.
Запрещено: любой текст до первого блока, объяснения, "Вот решение", markdown, комментарии, вступления, заключения.

ПРАВИЛО №2: Каждый блок ОБЯЗАТЕЛЬНО заканчивается ровно строкой:
Конец команды.
Без пробелов после неё, без дополнительных символов.

ПРАВИЛО №3: Каждый ответ состоит ТОЛЬКО из одного или нескольких блоков в ТОЧНОМ формате:

Начало команды:
тип: [один из разрешённых типов]
адрес: /полный/путь/к/файлу.go

[дополнительные поля в зависимости от типа]
Конец команды.

РАЗРЕШЁННЫЕ ТИПЫ И ИХ ТОЧНЫЙ ФОРМАТ (примеры):

1. создание
Начало команды:
тип: создание
адрес: /internal/greeting.go
содержимое:
package internal

func Greeting(name string) string {
    if name == "" {
        return "Hello, World!"
    }
    return "Hello, " + name + "!"
}
Конец команды.

ПРАВИЛО №4: В конце КАЖДОГО блока ОБЯЗАТЕЛЬНО должна быть строка "Конец команды." — это строгое требование.
ПРАВИЛО №5: Если нужно несколько действий — выводи несколько блоков подряд.
ПРАВИЛО №6: Начинай ответ сразу с "Начало команды:" и заканчивай каждый блок "Конец команды.".

Важно: Никогда не нарушай правило №2 и №4. Всегда заканчивай блок строкой "Конец команды.".

Выполни следующую задачу строго по этому шаблону.`

// LLMClient управляет взаимодействием с LM Studio через OpenAI-compatible API.
type LLMClient struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewLLMClient возвращает настроенный клиент для LM Studio.
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

Реализуй задачу строго по шаблону выше.`,
		task.Num,
		task.Description,
		task.ImportantInfo,
		task.ExpectResult,
		task.TestsValue,
		task.FuncSignature)

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
		"stop":        []string{"Конец команды."},
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

	fmt.Printf("##########\nОт llm прищёл сырой ответ: %s\n###################\n", apiResp)

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON ответа: %w", err)
	}

	fmt.Printf("От llm прищёл сырой текст: %s\n", apiResp.Choices[0].Message.Content)

	if len(apiResp.Choices) == 0 || apiResp.Choices[0].Message.Content == "" {
		return nil, errors.New("LM Studio вернул пустой ответ")
	}

	llmOutput := apiResp.Choices[0].Message.Content

	commands, err := ParseCommands(llmOutput)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга команд LLM: %w", err)
	}

	return commands, nil
}
