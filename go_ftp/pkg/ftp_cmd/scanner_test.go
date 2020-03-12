package ftp_cmd_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_cmd"
	"github.com/jakobsvenningsson/go_ftp/pkg/test_utils"
)

var tests = []struct {
	in          string
	expected    *ftp_cmd.Cmd
	expectedErr error
}{
	{"CWD file/path\n", &ftp_cmd.Cmd{ftp_cmd.CWD, "file/path"}, nil},
	{"CWD\n", nil, errors.New("No argument for command CWD")},
	{"CWD PWD\n", nil, errors.New("No argument for command CWD")},
	{"PWD\n", &ftp_cmd.Cmd{ftp_cmd.PWD, ""}, nil},
	{"USER demo\n", &ftp_cmd.Cmd{ftp_cmd.USER, "demo"}, nil},
	{"USER\n", nil, errors.New("No argument for command USER")},
	{"PASS pw\n", &ftp_cmd.Cmd{ftp_cmd.PASS, "pw"}, nil},
	{"PASS\n", nil, errors.New("No argument for command PASS")},
	{"RETR file\n", &ftp_cmd.Cmd{ftp_cmd.RETR, "file"}, nil},
	{"RETR\n", nil, errors.New("No argument for command RETR")},
	{"PORT 127.0.0.1:1234\n", &ftp_cmd.Cmd{ftp_cmd.PORT, "127.0.0.1:1234"}, nil},
	{"PORT\n", nil, errors.New("No argument for command PORT")},
	{"LIST\n", &ftp_cmd.Cmd{ftp_cmd.LIST, ""}, nil},
	{"PASV\n", &ftp_cmd.Cmd{ftp_cmd.PASV, ""}, nil},
	{"\n", nil, errors.New("No command")},
	{"", nil, errors.New("No command")},
}

func TestScanner(t *testing.T) {
	for _, test := range tests {
		scanner := ftp_cmd.NewScanner(strings.NewReader(test.in))
		cmd, err := scanner.NextCommand()
		if ok, have, want := test_utils.VerifyError(err, test.expectedErr); !ok {
			t.Errorf("Error actual = %v, and Expected = %v.", have, want)
		}
		// Verify Output
		switch {
		case cmd == nil && test.expected == nil:
		case cmd != nil && test.expected == nil:
			t.Errorf("Error actual = %v, and Expected = %v.", *cmd, nil)
		case cmd == nil && test.expected != nil:
			t.Errorf("Error actual = %v, and Expected = %v.", nil, test.expected)
		default:
			if *cmd != *test.expected {
				t.Errorf("Error actual = %v, and Expected = %v.", *cmd, *test.expected)
			}
		}
	}
}
