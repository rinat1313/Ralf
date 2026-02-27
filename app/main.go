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
		MaxTaskAttempts:       10,
		MaxCompileFixAttempts: 10,
		MaxTestAttempts:       10,
		WorkingDir:            ".",
	}
	if err := service.RunOrchestrator(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Оркестратор завершился ошибкой: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Все задачи обработаны успешно.")
}
