package client_connection

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_cmd"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_error"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_ip"
)

type ClientConnection struct {
	isAuth          bool
	user            string
	ip              string
	dirPath         ftpDirPath
	dataConn        dataConnection
	ctrlConn        io.ReadWriter
	ctrlConnScanner *ftp_cmd.Scanner
	authCh          chan AuthPkg
	mode            ftp_cmd.MODE
}

type dataConnection struct {
	addr string
	ch   chan dataPackage
	ln   net.Listener
	mode ftp_cmd.MODE
}

type AuthPkg struct {
	User     string
	Password string
	ReplyCh  chan bool
}

type dataPackage struct {
	action  byte
	payload []byte
	reply   chan int
}

// Public Methods

func New(conn io.ReadWriter, authCh chan AuthPkg, root, ip string) *ClientConnection {
	return &ClientConnection{
		isAuth:          false,
		ctrlConn:        conn,
		ctrlConnScanner: ftp_cmd.NewScanner(conn),
		authCh:          authCh,
		dirPath:         ftpDirPath{root, "/"},
		dataConn:        dataConnection{mode: ftp_cmd.PASSIVE},
		ip:              ip,
	}
}

func (cc *ClientConnection) Command() (*ftp_cmd.Cmd, error) {
	cmd, err := cc.ctrlConnScanner.NextCommand()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (cc *ClientConnection) Reply(cmd *ftp_cmd.Cmd) error {
	log.Printf("Replying to cmd %s, arg: %s.\n", cmd.Type, cmd.Arg)
	if !cc.isAuth && cc.needAuth(cmd) {
		err := cc.send(530, "Please login with USER and PASS.")
		return err
	}

	var err error
	switch cmd.Type {
	case ftp_cmd.USER:
		err = cc.handleUserCMD(cmd)
	case ftp_cmd.PASS:
		err = cc.handlePassCMD(cmd)
	case ftp_cmd.PWD:
		err = cc.handlePwdCMD(cmd)
	case ftp_cmd.CWD:
		err = cc.handleCwdCMD(cmd)
	case ftp_cmd.LIST:
		err = cc.handleListCMD(cmd)
	case ftp_cmd.EPSV:
		err = cc.handleEpsvCMD(cmd)
	case ftp_cmd.PASV:
		err = cc.handlePasvCMD(cmd)
	case ftp_cmd.PORT:
		err = cc.handlePortCMD(cmd)
	case ftp_cmd.RETR:
		err = cc.handleRetrCMD(cmd)
	case ftp_cmd.DELE:
		err = cc.handleDeleCMD(cmd)
	case ftp_cmd.STOR:
		err = cc.handleStorCMD(cmd)
	case ftp_cmd.TYPE:
		err = cc.notImplementedError(cmd)
	case ftp_cmd.QUIT:
		err = cc.send(221, "Goodbye")
		if conn, ok := cc.ctrlConn.(net.Conn); ok {
			conn.Close()
		}
	default:
		err = cc.send(500, fmt.Sprintf("'%s': command not understood.", cmd.Type))
	}
	return err
}

func (cc *ClientConnection) SendWelcomeMsg() error {
	return cc.send(220, "Service ready.")
}

// Private Methods

func (cc *ClientConnection) handleDeleCMD(cmd *ftp_cmd.Cmd) error {
	filepath, err := cc.getFilePathIfExist(cmd.Arg)
	if err != nil {
		return cc.send(550, "File not found.")
	}
	if err := os.Remove(filepath); err != nil {
		return err
	}
	return cc.send(200, "DELE command successful.")

}

func (cc *ClientConnection) handleStorCMD(cmd *ftp_cmd.Cmd) error {
	filePath := filepath.Join(cc.dirPath.path(), cmd.Arg)
	wait := make(chan struct{})
	cc.connectToDataConnection(func(ch chan dataPackage) {
		wait <- struct{}{}
		reply := make(chan int)
		buf := make([]byte, 8196)
		ch <- dataPackage{action: 'r', payload: buf, reply: reply}
		n := <-reply
		if err := ioutil.WriteFile(filePath, buf[:n], 0644); err != nil {
			log.Fatal(err)
		}
		wait <- struct{}{}
	})
	<-wait
	if err := cc.send(150, "Opening ASCII mode data connection for file."); err != nil {
		return err
	}
	<-wait
	return cc.send(226, "Transfer complete.")
}

func (cc *ClientConnection) handleUserCMD(cmd *ftp_cmd.Cmd) error {
	cc.user = cmd.Arg
	return cc.send(331, fmt.Sprintf("Password required for %s.", cc.user))
}

func (cc *ClientConnection) handlePassCMD(cmd *ftp_cmd.Cmd) error {
	replyCh := make(chan bool)
	cc.authCh <- AuthPkg{cc.user, cmd.Arg, replyCh}
	cc.isAuth = <-replyCh
	if !cc.isAuth {
		return cc.send(530, "Login failed.")
	}
	return cc.send(230, "User logged in.")
}

func (cc *ClientConnection) handlePwdCMD(cmd *ftp_cmd.Cmd) error {
	return cc.send(257, fmt.Sprintf("\"%s\" is current directory.", cc.dirPath.current))
}

func (cc *ClientConnection) handleCwdCMD(cmd *ftp_cmd.Cmd) error {
	path, err := cc.getFilePathIfExist(cmd.Arg)
	if err != nil {
		return cc.send(550, "Invalid path.")
	}
	p, err := filepath.Rel(cc.dirPath.root, path)
	if err != nil {
		log.Fatal(err)
	}
	cc.dirPath.current = filepath.Clean("/" + p)
	return cc.send(250, "CWD command successful.")
}

func (cc *ClientConnection) handleListCMD(cmd *ftp_cmd.Cmd) error {
	output, err := exec.Command("ls", "-l", cc.dirPath.path()).Output()
	if err != nil {
		return err
	}
	wait := make(chan struct{})
	cc.connectToDataConnection(func(ch chan dataPackage) {
		wait <- struct{}{}
		ch <- dataPackage{action: 'w', payload: output}
		wait <- struct{}{}
	})
	<-wait
	if err := cc.send(150, "Opening ASCII mode data connection for file list."); err != nil {
		return err
	}
	<-wait
	return cc.send(226, "Transfer complete.")
}

func (cc *ClientConnection) handleEpsvCMD(cmd *ftp_cmd.Cmd) error {
	ch, ln, port, err := cc.openDataListener()
	if err != nil {
		return err
	}
	cc.dataConn.ln = ln
	cc.dataConn.ch = ch
	cc.dataConn.mode = ftp_cmd.PASSIVE
	return cc.send(229, fmt.Sprintf("Entering Extended Passive Mode (|||%s|)).", port))
}

func (cc *ClientConnection) handlePasvCMD(cmd *ftp_cmd.Cmd) error {
	ch, ln, port, err := cc.openDataListener()
	if err != nil {
		return err
	}
	cc.dataConn.ln = ln
	cc.dataConn.ch = ch
	encoded, err := ftp_ip.Encode(cc.ip, port)
	if err != nil {
		return err
	}
	cc.mode = ftp_cmd.PASSIVE
	return cc.send(227, fmt.Sprintf("Entering Passive Mode (%s).", encoded))
}

func (cc *ClientConnection) handlePortCMD(cmd *ftp_cmd.Cmd) error {
	addr, err := ftp_ip.Decode(cmd.Arg)
	if err != nil {
		return err
	}
	if cc.dataConn.ln != nil {
		cc.dataConn.ln.Close()
		cc.dataConn.ln = nil
	}
	cc.dataConn.addr = addr
	cc.dataConn.mode = ftp_cmd.ACTIVE
	return cc.send(200, "PORT command successful.")
}

func (cc *ClientConnection) handleRetrCMD(cmd *ftp_cmd.Cmd) error {
	path, err := cc.getFilePathIfExist(cmd.Arg)
	if err != nil {
		return cc.send(550, "File not found.")
	}
	wait := make(chan struct{})
	cc.connectToDataConnection(func(ch chan dataPackage) {
		wait <- struct{}{}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			close(ch)
			log.Fatal(err)
		}
		ch <- dataPackage{action: 'w', payload: data}
		wait <- struct{}{}
	})
	<-wait
	err = cc.send(150, "Opening ASCII mode data connection.")
	if err != nil {
		return err
	}
	<-wait
	return cc.send(226, "Transfer complete.")

}

func (cc *ClientConnection) openDataListener() (chan dataPackage, net.Listener, string, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, "", err
	}
	ch := make(chan dataPackage)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			close(ch)
			return
		}
		defer conn.Close()
		for p := range ch {
			cc.handleDataPackage(conn, p)
		}
	}()
	tmp := strings.Split(listener.Addr().String(), ":")
	return ch, listener, tmp[len(tmp)-1], nil
}

func (cc *ClientConnection) openDataConnection(addr string) (chan dataPackage, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	ch := make(chan dataPackage)
	go func() {
		defer conn.Close()
		for p := range ch {
			cc.handleDataPackage(conn, p)
		}
	}()
	return ch, nil
}

func (cc *ClientConnection) needAuth(cmd *ftp_cmd.Cmd) bool {
	switch cmd.Type {
	case ftp_cmd.USER, ftp_cmd.PASS, ftp_cmd.QUIT:
		return false
	}
	return true
}

func (cc *ClientConnection) getDataChannel() (chan dataPackage, error) {
	switch cc.mode {
	case ftp_cmd.PASSIVE:
		return cc.dataConn.ch, nil
	case ftp_cmd.ACTIVE:
		return cc.openDataConnection(cc.dataConn.addr)
	default:
		return nil, errors.New("Invalid data transfer mode")
	}
}

func (cc *ClientConnection) notImplementedError(cmd *ftp_cmd.Cmd) error {
	err := cc.send(550, fmt.Sprintf("'%s': command not implemented.", cmd.Type))
	if err != nil {
		return err
	}
	return &ftp_error.NotImplementedError{string(cmd.Type)}
}

func (cc *ClientConnection) send(status int, text string) error {
	_, err := fmt.Fprintf(cc.ctrlConn, "%d %s\n", status, text)
	return err
}

func (cc *ClientConnection) getFilePathIfExist(fileName string) (string, error) {
	var newPath string
	if filepath.IsAbs(fileName) {
		newPath = fileName
	} else {
		newPath = filepath.Join(cc.dirPath.current, fileName)
	}
	filePath := filepath.Join(cc.dirPath.root, newPath)
	if !fileExist(filePath) {
		return "", &ftp_error.FileNotFoundError{filePath}
	}
	return filePath, nil
}

func (cc *ClientConnection) connectToDataConnection(action func(ch chan dataPackage)) {
	go func() {
		ch, err := cc.getDataChannel()
		if err != nil {
			log.Fatal(err)
		}
		defer close(ch)
		action(ch)
	}()
}

func (cc *ClientConnection) handleDataPackage(conn net.Conn, p dataPackage) {
	switch p.action {
	case 'r':
		n, err := conn.Read(p.payload)
		if err != nil {
			log.Fatal(err)
		}
		p.reply <- n
	case 'w':
		conn.Write(p.payload)
	}
}
