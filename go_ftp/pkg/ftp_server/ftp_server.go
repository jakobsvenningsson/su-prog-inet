package ftp_server

import (
	"fmt"
	"log"
	"net"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_error"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_server/client_connection"
)

type FtpServer struct {
	root      string
	port      string
	ip        string
	users     map[string]string
	usrAuthCh chan client_connection.AuthPkg
	listener  net.Listener
}

// Public Methods

func New(root, ip, port string) *FtpServer {
	return &FtpServer{
		root:      root,
		port:      port,
		ip:        ip,
		users:     map[string]string{"demo": "password"},
		usrAuthCh: make(chan client_connection.AuthPkg),
	}
}

func (ftpserver *FtpServer) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ftpserver.ip, ftpserver.port))
	if err != nil {
		return err
	}
	ftpserver.listener = ln
	go ftpserver.startAuthChannel()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		log.Printf("New connection accepted from %s.\n", conn.RemoteAddr())
		go ftpserver.handle(conn)
	}
	return nil
}

func (ftpserver *FtpServer) Stop() {
	ftpserver.listener.Close()
	close(ftpserver.usrAuthCh)
}

// Private Methods

func (ftpserver *FtpServer) handle(conn net.Conn) {
	cc := client_connection.New(conn, ftpserver.usrAuthCh, ftpserver.root, ftpserver.ip)
	if err := cc.SendWelcomeMsg(); err != nil {
		log.Fatal(err)
	}
Loop:
	for {
		cmd, err := cc.Command()
		if err != nil {
			fmt.Println(err.Error())
			switch err.(type) {
			case *ftp_error.NotImplementedError:
				continue
			default:
				break Loop
			}
		}
		if err := cc.Reply(cmd); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
	log.Printf("Connection closed %s.\n", conn.RemoteAddr())

}

func (ftpserver *FtpServer) startAuthChannel() {
	for authPkg := range ftpserver.usrAuthCh {
		pw, ok := ftpserver.users[authPkg.User]
		if !ok {
			authPkg.ReplyCh <- false
		}
		authPkg.ReplyCh <- pw == authPkg.Password
	}
}
