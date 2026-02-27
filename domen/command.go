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
	Type    string            `json:"Type"`
	Path    string            `json:"Path"`
	Content string            `json:"Content"`
	Lines   map[string]string `json:"Lines"`
	SrcPath string            `json:"SrcPath"`
	DstPath string            `json:"DstPath"`
}
