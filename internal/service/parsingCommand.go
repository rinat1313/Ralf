package service

import (
	"Ralf/domen"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseCommands выполняет полный парсинг ответа от LM Studio в слайс команд.
func ParseCommands(response string) ([]domen.Command, error) {
	var commands []domen.Command

	fmt.Printf("От LLM получен ответ %s\n", response)

	blockRe := regexp.MustCompile(`(?s)Начало команды:(.*?)(?:Конец команды\.|$)`)
	matches := blockRe.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		block := strings.TrimSpace(match[1])

		cmd, err := parseSingleCommandBlock(block)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга блока: %w", err)
		}
		commands = append(commands, cmd)
	}

	if len(commands) == 0 {
		return nil, errors.New("в ответе LM Studio не обнаружено ни одного блока команды")
	}

	return commands, nil
}

// parseSingleCommandBlock парсит один блок "Начало команды: ...".
func parseSingleCommandBlock(block string) (domen.Command, error) {
	cmd := domen.Command{
		Lines: make(map[int]string),
	}

	// Тип команды
	if m := regexp.MustCompile(`тип:\s*([^\n]+)`).FindStringSubmatch(block); len(m) > 1 {
		typStr := strings.TrimSpace(m[1])
		if t, ok := mapRussianType(typStr); ok {
			cmd.Type = t
		} else {
			cmd.Type = domen.CommandType(typStr)
		}
	} else {
		return cmd, errors.New("тип команды не найден")
	}

	// Основной путь
	if m := regexp.MustCompile(`адрес:\s*([^\n]+)`).FindStringSubmatch(block); len(m) > 1 {
		cmd.Path = strings.TrimSpace(m[1])
	}

	// Исходный и целевой пути
	if m := regexp.MustCompile(`исходный_адрес:\s*([^\n]+)`).FindStringSubmatch(block); len(m) > 1 {
		cmd.SrcPath = strings.TrimSpace(m[1])
	}
	if m := regexp.MustCompile(`целевой_адрес:\s*([^\n]+)`).FindStringSubmatch(block); len(m) > 1 {
		cmd.DstPath = strings.TrimSpace(m[1])
	}

	// Содержимое файла (создание)
	if cmd.Type == domen.CmdCreate {
		re := regexp.MustCompile(`(?s)содержимое:\s*[\w]*\n(.*?)\n`)
		if m := re.FindStringSubmatch(block); len(m) > 1 {
			cmd.Content = strings.TrimSpace(m[1])
		}
	}

	// Строки для редактирования или добавления новых строк
	if cmd.Type == domen.CmdEdit || cmd.Type == domen.CmdAddLines {
		re := regexp.MustCompile(`(?s)строки:\s*\{(.*?)\}`)
		if m := re.FindStringSubmatch(block); len(m) > 1 {
			cmd.Lines = parseLinesMap(m[1])
		}
	}

	return cmd, nil
}

// mapRussianType преобразует русское название типа в CommandType.
func mapRussianType(typStr string) (domen.CommandType, bool) {
	switch typStr {
	case "создание":
		return domen.CmdCreate, true
	case "удаление":
		return domen.CmdDelete, true
	case "внесение изменений":
		return domen.CmdEdit, true
	case "добавление строк":
		return domen.CmdAddLines, true
	case "копирование":
		return domen.CmdCopy, true
	case "перемещение":
		return domen.CmdMove, true
	case "чтение":
		return domen.CmdRead, true
	case "компиляция":
		return domen.CmdCompileCode, true
	}
	return "", false
}

// parseLinesMap парсит содержимое {10:"text", 20:"text2"} в map.
func parseLinesMap(s string) map[int]string {
	m := make(map[int]string)
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, ":"); idx != -1 {
			keyStr := strings.TrimSpace(p[:idx])
			value := strings.Trim(strings.TrimSpace(p[idx+1:]), `"`)
			if num, err := strconv.Atoi(keyStr); err == nil {
				m[num] = value
			}
		}
	}
	return m
}
