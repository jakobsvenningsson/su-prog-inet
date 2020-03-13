package ftp_cmd

type CmdType string

const (
	LIST CmdType = "LIST"
	USER         = "USER"
	PASS         = "PASS"
	RETR         = "RETR"
	PWD          = "PWD"
	CWD          = "CWD"
	PASV         = "PASV"
	PORT         = "PORT"
	QUIT         = "QUIT"
	EPSV         = "EPSV"
	TYPE         = "TYPE"
)

var cmds = []CmdType{
	LIST,
	USER,
	PASS,
	RETR,
	PWD,
	CWD,
	PASV,
	PORT,
	QUIT,
	EPSV,
	TYPE,
}

func (cmd CmdType) IsDataCMD() bool {
	switch cmd {
	case LIST, RETR:
		return true
	}
	return false
}

func HasArg(cmd CmdType) bool {
	switch cmd {
	case RETR, PASS, USER, CWD, PORT, TYPE:
		return true
	}
	return false
}

func isCommand(str string) bool {
	for _, cmd := range cmds {
		if string(cmd) == str {
			return true
		}
	}
	return false
}
