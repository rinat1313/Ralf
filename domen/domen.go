package domen

// TaskStatus представляет допустимые статусы задачи.
type TaskStatus string

const (
	StatusNew   TaskStatus = "new"   // задача готова к выполнению
	StatusRun   TaskStatus = "run"   // задача выполняется
	StatusError TaskStatus = "error" // при выполнении произошла ошибка
	StatusOK    TaskStatus = "ok"    // задача успешно выполнена
)

// Task описывает структуру задачи, прочитанной из файла.
type Task struct {
	Num           int        // номер задачи
	Description   string     // описание задачи
	ImportantInfo string     // важные моменты
	ExpectResult  string     // ожидаемый результат
	TestsValue    string     // тестовые данные
	FuncSignature string     // сигнатура функции (может быть пустой)
	Status        TaskStatus // текущий статус
}
