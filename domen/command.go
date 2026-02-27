package domen

// CommandType определяет допустимые типы команд от LM Studio.
type CommandType string

const (
	CmdCreate      CommandType = "создание"
	CmdDelete      CommandType = "удаление"
	CmdEdit        CommandType = "внесение изменений"
	CmdCopy        CommandType = "копирование"
	CmdMove        CommandType = "перемещение"
	CmdRead        CommandType = "чтение"
	CmdAddLines    CommandType = "добавление строк"
	CmdDeleteLines CommandType = "удаление строк"
	CmdCompileCode CommandType = "компиляция"
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
