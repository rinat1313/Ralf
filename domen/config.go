package domen

type Config struct {
	TasksFilePath         string // путь к файлу задач
	MaxTaskAttempts       int    // максимум попыток на одну задачу (общий цикл)
	MaxCompileFixAttempts int    // максимум циклов исправления компиляции
	MaxTestAttempts       int    // максимум попыток генерации тестов
	WorkingDir            string // рабочая директория проекта
}
