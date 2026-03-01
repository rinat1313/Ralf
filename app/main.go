package main

import (
	"Ralf/domen"
	"Ralf/internal/service"
	"fmt"
	"os"
)

func main() {
	cfg := domen.Config{
		TasksFilePath:         "tasks.txt",
		MaxTaskAttempts:       5,
		MaxCompileFixAttempts: 5,
		MaxTestAttempts:       5,
		WorkingDir:            ".",
	}
	if err := service.RunOrchestrator(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Оркестратор завершился ошибкой: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Все задачи обработаны успешно.")
}
