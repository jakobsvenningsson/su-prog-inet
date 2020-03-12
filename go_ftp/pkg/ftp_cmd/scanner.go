package ftp_cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	scanner.Split(bufio.ScanWords)
	return &Scanner{
		in: scanner,
	}
}

func (p *Scanner) NextCommand() (*Cmd, error) {
	word, ok := p.nextWord()
	if !ok {
		return nil, errors.New("No command")
	}
	cmd := CmdType(word)
	if !HasArg(cmd) {
		return &Cmd{Type: cmd}, nil
	}
	arg, ok := p.nextWord()
	if !ok || isCommand(arg) {
		return nil, fmt.Errorf("No argument for command %s", cmd)
	}
	return &Cmd{cmd, arg}, nil
}

func (p *Scanner) nextWord() (string, bool) {
	if !p.in.Scan() {
		return "", false
	}
	return p.in.Text(), true
}

func (p *Scanner) hasNext() bool {
	return p.in.Scan()
}
