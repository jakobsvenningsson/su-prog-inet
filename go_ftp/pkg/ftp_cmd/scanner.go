package ftp_cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Cmd struct {
	Type CmdType
	Arg  string
}

type Scanner struct {
	in *bufio.Scanner
}

func NewScanner(input io.Reader) *Scanner {
	scanner := bufio.NewScanner(input)
	//scanner.Split(bufio.ScanWords)
	return &Scanner{
		in: scanner,
	}
}

func (p *Scanner) NextCommand() (*Cmd, error) {
	line, ok := p.nextLine()
	if !ok {
		return nil, errors.New("No command")
	}
	components := strings.SplitN(line, " ", 2)
	word := components[0]

	if !isCommand(word) {
		return nil, errors.New("Invalid command")
	}
	cmd := CmdType(word)
	if !HasArg(cmd) {
		return &Cmd{Type: cmd}, nil
	}
	//arg, ok := p.nextWord()
	if len(components) < 2 || isCommand(components[1]) {
		return nil, fmt.Errorf("No argument for command %s", cmd)
	}
	return &Cmd{cmd, components[1]}, nil
}

func (p *Scanner) nextLine() (string, bool) {
	if !p.in.Scan() {
		return "", false
	}
	return p.in.Text(), true
}

func (p *Scanner) hasNext() bool {
	return p.in.Scan()
}
