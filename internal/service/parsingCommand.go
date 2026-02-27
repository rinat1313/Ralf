package service

import (
	"Ralf/domen"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseCommands теперь парсит чистый JSON-массив напрямую в []domen.Command.
// domen.Command уже имеет поля и json-теги, соответствующие выводу LLM.
func ParseCommands(response string) ([]domen.Command, error) {
	fmt.Printf("От LLM получен JSON-ответ: %s\n", response)

	// Убираем возможные markdown-обёртки (```json ... ```)
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		if len(lines) > 1 {
			response = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	var commands []domen.Command
	if err := json.Unmarshal([]byte(response), &commands); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if len(commands) == 0 {
		return nil, errors.New("в ответе LM Studio не обнаружено ни одной команды")
	}

	return commands, nil
}
