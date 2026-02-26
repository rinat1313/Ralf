package service

import (
	"Ralf/domen"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExecuteCommand выполняет переданную команду.
// Возвращает содержимое файла только для CmdRead, иначе пустую строку.
func ExecuteCommand(cmd domen.Command) (string, error) {
	switch cmd.Type {
	case domen.CmdCreate:
		return "", executeCreate(cmd)
	case domen.CmdDelete:
		return "", executeDelete(cmd)
	case domen.CmdEdit:
		return "", executeEdit(cmd)
	case domen.CmdAddLines:
		return "", executeAddLines(cmd)
	case domen.CmdDeleteLines:
		return "", executeDeleteLines(cmd)
	case domen.CmdCopy:
		return "", executeCopy(cmd)
	case domen.CmdMove:
		return "", executeMove(cmd)
	case domen.CmdRead:
		return executeRead(cmd)
	default:
		return "", fmt.Errorf("неизвестный тип команды: %s", cmd.Type)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func executeCreate(cmd domen.Command) error {
	if fileExists(cmd.Path) {
		return fmt.Errorf("файл уже существует: %s", cmd.Path)
	}
	if cmd.Content == "" {
		return errors.New("пустое содержимое для создания файла")
	}
	if dir := filepath.Dir(cmd.Path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию: %w", err)
		}
	}
	return os.WriteFile(cmd.Path, []byte(cmd.Content), 0644)
}

func executeDelete(cmd domen.Command) error {
	if !fileExists(cmd.Path) {
		return fmt.Errorf("файл не существует: %s", cmd.Path)
	}
	return os.Remove(cmd.Path)
}

func executeEdit(cmd domen.Command) error {
	if !fileExists(cmd.Path) {
		return fmt.Errorf("файл не существует: %s", cmd.Path)
	}
	if len(cmd.Lines) == 0 {
		return errors.New("нет строк для изменения")
	}
	lines, err := readLines(cmd.Path)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл %s: %w", cmd.Path, err)
	}
	for lineNum, newText := range cmd.Lines {
		if lineNum < 1 || lineNum > len(lines) {
			return fmt.Errorf("строка %d не существует в файле %s (всего строк: %d)", lineNum, cmd.Path, len(lines))
		}
		lines[lineNum-1] = newText
	}
	return writeLines(cmd.Path, lines)
}

func executeAddLines(cmd domen.Command) error {
	if !fileExists(cmd.Path) {
		return fmt.Errorf("файл не существует: %s", cmd.Path)
	}
	if len(cmd.Lines) == 0 {
		return errors.New("нет строк для добавления")
	}
	lines, err := readLines(cmd.Path)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл %s: %w", cmd.Path, err)
	}
	currentLen := len(lines)
	keys := make([]int, 0, len(cmd.Lines))
	for k := range cmd.Lines {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	if keys[0] != currentLen+1 {
		return fmt.Errorf("нельзя добавить строку %d: файл содержит только %d строк", keys[0], currentLen)
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] != keys[i-1]+1 {
			return errors.New("строки для добавления должны идти последовательно без пропусков")
		}
	}
	for _, k := range keys {
		lines = append(lines, cmd.Lines[k])
	}
	return writeLines(cmd.Path, lines)
}

func executeDeleteLines(cmd domen.Command) error {
	if !fileExists(cmd.Path) {
		return fmt.Errorf("файл не существует: %s", cmd.Path)
	}
	if len(cmd.Lines) == 0 {
		return errors.New("нет строк для удаления")
	}
	lines, err := readLines(cmd.Path)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл %s: %w", cmd.Path, err)
	}
	keys := make([]int, 0, len(cmd.Lines))
	for k := range cmd.Lines {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys))) // удаляем с большей строки к меньшей
	for _, k := range keys {
		if k < 1 || k > len(lines) {
			return fmt.Errorf("строка %d не существует в файле %s (всего строк: %d)", k, cmd.Path, len(lines))
		}
		idx := k - 1
		lines = append(lines[:idx], lines[idx+1:]...)
	}
	return writeLines(cmd.Path, lines)
}

func executeCopy(cmd domen.Command) error {
	if cmd.SrcPath == "" || cmd.DstPath == "" {
		return errors.New("не указаны пути для копирования")
	}
	if !fileExists(cmd.SrcPath) {
		return fmt.Errorf("исходный файл не существует: %s", cmd.SrcPath)
	}
	if fileExists(cmd.DstPath) {
		return fmt.Errorf("целевой файл уже существует: %s", cmd.DstPath)
	}
	if dir := filepath.Dir(cmd.DstPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию: %w", err)
		}
	}
	data, err := os.ReadFile(cmd.SrcPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать исходный файл: %w", err)
	}
	return os.WriteFile(cmd.DstPath, data, 0644)
}

func executeMove(cmd domen.Command) error {
	if cmd.SrcPath == "" || cmd.DstPath == "" {
		return errors.New("не указаны пути для перемещения")
	}
	if !fileExists(cmd.SrcPath) {
		return fmt.Errorf("исходный файл не существует: %s", cmd.SrcPath)
	}
	if fileExists(cmd.DstPath) {
		return fmt.Errorf("целевой файл уже существует: %s", cmd.DstPath)
	}
	if dir := filepath.Dir(cmd.DstPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию: %w", err)
		}
	}
	return os.Rename(cmd.SrcPath, cmd.DstPath)
}

func executeRead(cmd domen.Command) (string, error) {
	if !fileExists(cmd.Path) {
		return "", fmt.Errorf("файл не существует: %s", cmd.Path)
	}
	data, err := os.ReadFile(cmd.Path)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл %s: %w", cmd.Path, err)
	}
	return string(data), nil
}
