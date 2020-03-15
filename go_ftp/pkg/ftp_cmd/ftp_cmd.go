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
	DELE         = "DELE"
	STOR         = "STOR"
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
	DELE,
	STOR,
}

func (cmd CmdType) IsDataCMD() bool {
	switch cmd {
	case LIST, RETR, STOR:
		return true
	}
	return false
}

func HasArg(cmd CmdType) bool {
	switch cmd {
	case RETR, PASS, USER, CWD, PORT, TYPE, STOR, DELE:
		return true
	}
	return false
}

func IsCommand(str string) bool {
	for _, cmd := range cmds {
		if string(cmd) == str {
			return true
		}
	}
	return false
}
