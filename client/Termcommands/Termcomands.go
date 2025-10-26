package termcommands

type TermComands string

const (
	Clear            TermComands = "\033[H\033[2J"
	SaveCursorDEC    TermComands = "\033 7"
	RestoreCursorDEC TermComands = "\033 8"
	SaveCursorSCO    TermComands = "\033[s"
	RestoreCursorSCO TermComands = "\033[u"
	Save             TermComands = "\033[?47h" // Not all suppport
	Restore          TermComands = "\033[?47l" // Not all suppport
	ClearLine        TermComands = "\033[2K\r"
	LineUp           TermComands = "\033[F"
	OwnTextColour    TermComands = "\033[38;5;46;48;5;0m"
	OtherTextColour  TermComands = "\033[38;5;45;48;5;0m"
	ServerTextColour TermComands = "\033[38;5;214;48;5;0m"
	ResetColour      TermComands = "\033[39m\033[49m"
)
