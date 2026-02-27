package service

import (
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Compile выполняет компиляцию Go-кода по указанному пути (файл или директория проекта)
// после применения атомарных команд Command. Использует go build для выявления
// максимального количества критических ошибок компиляции (синтаксис, типы,
// неиспользуемые идентификаторы и т.д.).
func Compile(path string) (string, error) {
	// Определяем рабочую директорию
	dir := filepath.Dir(path)
	if dir == "." || dir == "" || dir == "/" {
		dir = path
	}

	// Кросс-платформенный путь к /dev/null
	devNull := "/dev/null"
	if runtime.GOOS == "windows" {
		devNull = "NUL"
	}

	// Формируем аргументы: для .go-файла собираем только его,
	// иначе весь пакет и подпакеты
	args := []string{"build", "-o", devNull, "./..."}
	if strings.HasSuffix(strings.ToLower(path), ".go") {
		args = []string{"build", "-o", devNull, path}
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	// Захватываем весь вывод (stdout + stderr)
	output, err := cmd.CombinedOutput()

	if err != nil {
		compileLog := string(output)
		if compileLog == "" {
			compileLog = err.Error()
		}
		return compileLog, errors.New("Ошибка компиляции.")
	}

	return "", nil
}
