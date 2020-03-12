package client_connection

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_cmd"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_ip"
)

type ClientConnection struct {
	isAuth   bool
	user     string
	password string
	path     string
	root     string
	ip       string
	conn     io.ReadWriter
	scanner  *ftp_cmd.Scanner
	authCh   chan AuthPkg
	dataChs  []chan []byte
	mutex    sync.Mutex
}

type AuthPkg struct {
	User     string
	Password string
	ReplyCh  chan bool
}

// Public Methods

func New(conn io.ReadWriter, authCh chan AuthPkg, root, ip string) *ClientConnection {
	cc := &ClientConnection{
		isAuth:  false,
		conn:    conn,
		scanner: ftp_cmd.NewScanner(conn),
		authCh:  authCh,
		path:    "/",
		root:    root,
		dataChs: [](chan []byte){},
		ip:      ip,
	}

	return cc
}

func (cc *ClientConnection) Command() (*ftp_cmd.Cmd, error) {
	token, err := cc.scanner.NextCommand()
	if err != nil {
		cc.send(500, "Error parsing command.")
		return nil, err
	}
	return token, nil
}

func (cc *ClientConnection) Reply(token *ftp_cmd.Cmd) (bool, error) {
	log.Printf("Replying token %s %s.\n", token.Type, token.Arg)
	if !cc.isAuth && (token.Type != ftp_cmd.USER && token.Type != ftp_cmd.PASS) {
		err := cc.send(530, "Please login with USER and PASS.")
		return false, err
	}

	var err error
	switch token.Type {
	case ftp_cmd.USER:
		err = cc.handleUserCMD(token)
	case ftp_cmd.PASS:
		err = cc.handlePassCMD(token)
	case ftp_cmd.PWD:
		err = cc.handlePwdCMD(token)
	case ftp_cmd.CWD:
		err = cc.handleCwdCMD(token)
	case ftp_cmd.LIST:
		err = cc.handleListCMD(token)
	case ftp_cmd.PASV:
		err = cc.handlePasvCMD(token)
	case ftp_cmd.PORT:
		err = cc.handlePortCMD(token)
	case ftp_cmd.RETR:
		err = cc.handleRetrCMD(token)
	case ftp_cmd.QUIT:
		return true, nil
	default:
		if err := cc.send(500, fmt.Sprintf("'%s': command not understood.", token.Type)); err != nil {
			return false, err
		}
	}
	return false, err
}

func (cc *ClientConnection) SendWelcomeMsg() error {
	if err := cc.send(220, "Service ready."); err != nil {
		return err
	}
	return nil
}

// Private Methods

func (cc *ClientConnection) send(status int, text string) error {
	msg := fmt.Sprintf("%d %s\n", status, text)
	_, err := cc.conn.Write([]byte(msg))
	return err
}

func (cc *ClientConnection) handleUserCMD(token *ftp_cmd.Cmd) error {
	cc.user = token.Arg
	return cc.send(331, fmt.Sprintf("Password required for %s.", cc.user))
}

func (cc *ClientConnection) handlePassCMD(token *ftp_cmd.Cmd) error {
	cc.password = token.Arg
	replyCh := make(chan bool)
	cc.authCh <- AuthPkg{cc.user, cc.password, replyCh}
	isAuth := <-replyCh
	if !isAuth {
		return cc.send(530, "Login failed.")
	}
	cc.isAuth = true
	return cc.send(230, "User logged in.")
}

func (cc *ClientConnection) handlePwdCMD(token *ftp_cmd.Cmd) error {
	return cc.send(257, cc.path+" is current directory.")
}

func (cc *ClientConnection) handleCwdCMD(token *ftp_cmd.Cmd) error {
	path := token.Arg
	var newPath string
	if filepath.IsAbs(path) {
		newPath = path
	} else {
		newPath = filepath.Join(cc.path, path)
	}
	if !cc.dirExist(newPath) {
		return cc.send(550, "Invalid path.")
	}
	cc.path = filepath.Clean(newPath)
	return cc.send(250, "CWD command successful.")
}

func (cc *ClientConnection) handleListCMD(token *ftp_cmd.Cmd) error {
	cc.mutex.Lock()
	ch := cc.dataChs[0]
	cc.dataChs = cc.dataChs[1:]
	cc.mutex.Unlock()

	output, err := exec.Command("ls", "-l", cc.root+cc.path).Output()
	if err != nil {
		return err
	}
	ch <- output
	close(ch)
	return cc.send(150, "Opening ASCII mode data connection for file list.")
}

func (cc *ClientConnection) handlePasvCMD(token *ftp_cmd.Cmd) error {
	ch, port, err := cc.openDataConnection()
	if err != nil {
		return err
	}
	cc.mutex.Lock()
	cc.dataChs = append(cc.dataChs, ch)
	cc.mutex.Unlock()
	encoded, err := ftp_ip.Encode(fmt.Sprintf("%s:%s", cc.ip, port))
	if err != nil {
		return err
	}
	return cc.send(227, encoded)
}

func (cc *ClientConnection) handlePortCMD(token *ftp_cmd.Cmd) error {
	//"200 PORT command successful."
	return cc.send(202, "Command not implemented.")
}

func (cc *ClientConnection) handleRetrCMD(token *ftp_cmd.Cmd) error {
	file := token.Arg
	var path string
	if filepath.IsAbs(file) {
		path = filepath.Join(cc.root, file)
	} else {
		path = filepath.Join(cc.root, cc.path, file)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cc.send(550, "File not found.")
	}

	go func() {
		cc.mutex.Lock()
		ch := cc.dataChs[0]
		cc.dataChs = cc.dataChs[1:]
		cc.mutex.Unlock()
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		ch <- data
		close(ch)
	}()

	return cc.send(150, "Opening ASCII mode data connection.")
}

func (cc *ClientConnection) openDataConnection() (chan []byte, string, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, "", err
	}
	ch := make(chan []byte)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		for data := range ch {
			conn.Write(data)
		}
	}()
	tmp := strings.Split(listener.Addr().String(), ":")
	return ch, tmp[len(tmp)-1], nil
}

func (cc *ClientConnection) dirExist(path string) bool {
	if src, err := os.Stat(filepath.Join(cc.root, path)); !os.IsNotExist(err) {
		return src.IsDir()
	}
	return false
}
