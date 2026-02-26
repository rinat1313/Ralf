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

ПРАВИЛО №2: Каждый ответ состоит ТОЛЬКО из одного или нескольких блоков в ТОЧНОМ формате:

Начало команды:
тип: [один из разрешённых типов]
адрес: /полный/путь/к/файлу.go

[дополнительные поля в зависимости от типа]
Конец команды.

РАЗРЕШЁННЫЕ ТИПЫ И ИХ ТОЧНЫЙ ФОРМАТ:

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

2. внесение изменений
Начало команды:
тип: внесение изменений
адрес: /main.go
строки:
{
1:"package main",
10:"    fmt.Println(Greeting(\"Alice\"))",
}
Конец команды.

3. добавление строк
Начало команды:
тип: добавление строк
адрес: /main.go
строки:
{
15:"func main() {",
16:"    fmt.Println(\"Done\")",
17:"}",
}
Конец команды.

4. удаление строк
Начало команды:
тип: удаление строк
адрес: /old.go
строки:
{
5:"",
8:"",
}
Конец команды.

5. удаление
Начало команды:
тип: удаление
адрес: /obsolete.go
Конец команды.

6. копирование
Начало команды:
тип: копирование
адрес: /new/path.go
исходный_адрес: /old/path.go
целевой_адрес: /new/path.go
Конец команды.

7. перемещение
Начало команды:
тип: перемещение
адрес: /new/path.go
исходный_адрес: /old/path.go
целевой_адрес: /new/path.go
Конец команды.

8. чтение
Начало команды:
тип: чтение
адрес: /file.go
Конец команды.

ПРАВИЛО №3: Если нужно выполнить несколько действий — выводи несколько блоков подряд без разделителей.
ПРАВИЛО №4: Начинай ответ сразу с "Начало команды:" и заканчивай каждый блок "Конец команды.".
ПРАВИЛО №5: Никогда не нарушай этот формат ни при каких обстоятельствах.

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

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON ответа: %w", err)
	}

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
