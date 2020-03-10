package parser

import (
	"bufio"
	"errors"
	"io"

	"../ftpcommand"
	"../ftp_ip"

)

type Token struct {
	Type ftpcommand.CMD
	Arg  string
}

type Parser struct {
	in *bufio.Scanner
}

func New(input io.Reader) *Parser {
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanWords)
	return &Parser{
		in: scanner,
	}
}

func (p *Parser) NextToken() (*Token, error) {
	word, ok := p.nextWord()
	if !ok {
		return nil, errors.New("No token")
	}
	cmd := ftpcommand.CMD(word)
	if !ftpcommand.HasArg(cmd) {
		return &Token{Type: cmd}, nil
	}
	arg, ok := p.nextWord()
	if cmd == ftpcommand.PORT {
		arg = ftp_ip.Encode(arg)
	}
	if !ok {
		return nil, errors.New("No argument")
	}
	return &Token{cmd, arg}, nil

}

func (p *Parser) nextWord() (string, bool) {
	if !p.in.Scan() {
		return "", false
	}
	return p.in.Text(), true
}

func (p *Parser) hasNext() bool {
	return p.in.Scan()
}
