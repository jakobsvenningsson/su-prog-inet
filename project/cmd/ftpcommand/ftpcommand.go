package ftpcommand

type CMD string

const (
	LIST CMD = "LIST"
	USER     = "USER"
	PASS     = "PASS"
	RETR     = "RETR"
	PWD      = "PWD"
	CWD      = "CWD"
	PASV	 = "PASV"
	PORT     = "PORT"
)

func (cmd CMD) IsDataCMD() bool {
	switch cmd {
	case LIST, RETR:
		return true
	}
	return false
}

func HasArg(cmd CMD) bool {
	switch cmd {
	case RETR, PASS, USER, CWD, PORT:
		return true
	}
	return false
}

type MODE int

const (
	ACTIVE MODE = iota
	PASSIVE
)
