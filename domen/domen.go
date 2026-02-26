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

// CommandType определяет допустимые типы команд от LM Studio.
type CommandType string

const (
	CmdCreate   CommandType = "создание"
	CmdDelete   CommandType = "удаление"
	CmdEdit     CommandType = "внесение изменений"
	CmdCopy     CommandType = "копирование"
	CmdMove     CommandType = "перемещение"
	CmdRead     CommandType = "чтение"
	CmdAddLines CommandType = "добавление строк"
)

// Command представляет одну атомарную команду для модификации файловой системы.
type Command struct {
	Type    CommandType
	Path    string         // основной путь к файлу
	Content string         // полный контент файла (для создания)
	Lines   map[int]string // изменения по номерам строк (для редактирования)
	SrcPath string         // исходный путь (копирование/перемещение)
	DstPath string         // целевой путь (копирование/перемещение)
}
