package ftp_client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_cmd"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_error"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_ip"
)

type FtpClient struct {
	ctrlConn        io.ReadWriter
	ctrlConnScanner *bufio.Scanner
	usrIn           *ftp_cmd.Scanner
	connectionMode  ftp_cmd.MODE
	dataConnAddr    string
	outDir          string
	ioCh            chan string
}

// Public Methods

func New(usrIn io.Reader, ctrlConn io.ReadWriter, dataConnAddr, outDir string) (*FtpClient, error) {
	return &FtpClient{
		ctrlConn:        ctrlConn,
		ctrlConnScanner: bufio.NewScanner(ctrlConn),
		usrIn:           ftp_cmd.NewScanner(usrIn),
		connectionMode:  ftp_cmd.NOT_SET,
		dataConnAddr:    dataConnAddr,
		ioCh:            make(chan string),
		outDir:          outDir,
	}, nil
}

func (client *FtpClient) ProcessCommands() error {
	// This WaitGroup is used to make sure that all IO have been printed before the client exits.
	var ioWg sync.WaitGroup
	ioWg.Add(1)
	defer ioWg.Wait()
	defer close(client.ioCh)
	// Start IO goroutine.
	go func() {
		for s := range client.ioCh {
			fmt.Printf(s)
		}
		ioWg.Done()
	}()
	// This WaitGroup is used to make sure that all data transfers have finished before the client exits.
	var dataWg sync.WaitGroup
	defer dataWg.Wait()

Loop:
	for {
		cmd, err := client.usrIn.NextCommand()
		if err != nil {
			log.Printf("%s.\n", err.Error())
			switch err.(type) {
			case *ftp_error.NoArgumentError, *ftp_error.InvalidCommandError:
				continue
			default:
				break Loop
			}
		}
		log.Printf("Processing cmd: %s, arg: %s.\n", cmd.Type, cmd.Arg)
		status, reply, err := client.processCommand(cmd, &dataWg)
		if err != nil {
			log.Printf("%s.\n", err.Error())
			switch err.(type) {
			case *ftp_error.ExitError:
				break Loop
			default:
				continue
			}
		}
		client.ioCh <- fmt.Sprintf("%d %s\n", status, reply)
	}
	return nil
}

func (client *FtpClient) Authenticate(user, pw string) error {
	// 1. Send user using the "USER :user" FTP command
	status, _, err := client.processCommand(&ftp_cmd.Cmd{ftp_cmd.USER, user}, nil)
	if err != nil {
		return err
	}
	if status != 331 {
		return unexpectedStatusError(status, 331)
	}
	// 2. Send password using the "PASS :password" FTP command
	status, _, err = client.processCommand(&ftp_cmd.Cmd{ftp_cmd.PASS, pw}, nil)
	if err != nil {
		return err
	}
	if status != 230 {
		return unexpectedStatusError(status, 230)
	}
	return nil
}

func (client *FtpClient) ReadWelcomeMessage() (int, string, error) {
	return client.readCtrlConn()
}

// Private Methods

func (client *FtpClient) processCommand(command *ftp_cmd.Cmd, wg *sync.WaitGroup) (int, string, error) {
	cmd, arg := command.Type, command.Arg
	switch cmd {
	case ftp_cmd.LIST, ftp_cmd.RETR, ftp_cmd.STOR:
		if client.connectionMode == ftp_cmd.NOT_SET {
			return 0, "", errors.New("No connection mode specified")
		}
		go client.startDataConnection(cmd, arg, client.dataConnAddr, wg)
	case ftp_cmd.PORT:
		tmp := strings.Split(arg, ":")
		encodedArg, err := ftp_ip.Encode(tmp[0], tmp[1])
		if err != nil {
			return 0, "", err
		}
		client.connectionMode = ftp_cmd.ACTIVE
		client.dataConnAddr = arg
		arg = encodedArg
	default:
		if !ftp_cmd.IsCommand(string(cmd)) {
			return 0, "", &ftp_error.NotImplementedError{string(cmd)}
		}
	}

	if _, err := client.send(cmd, arg); err != nil {
		return 0, "", err
	}
	return client.handleReply(cmd)
}

func (client *FtpClient) handleReply(cmd ftp_cmd.CmdType) (int, string, error) {
	status, reply, err := client.readReply()
	if err != nil {
		return 0, "", err
	}
	switch cmd {
	case ftp_cmd.PASV:
		addr, err := ftp_ip.Decode(reply)
		if err != nil {
			return 0, "", err
		}
		client.dataConnAddr = addr
		client.connectionMode = ftp_cmd.PASSIVE
	case ftp_cmd.QUIT:
		return status, reply, &ftp_error.ExitError{}
	case ftp_cmd.PORT:
		client.connectionMode = ftp_cmd.ACTIVE
	default:
	}
	return status, reply, nil
}

func (client *FtpClient) startDataConnection(cmd ftp_cmd.CmdType, arg, addr string, wg *sync.WaitGroup) error {
	wg.Add(1)
	log.Printf("Starting data channel on addr %s.\n", addr)
	dataAction, err := getDataAction(cmd)
	if err != nil {
		return err
	}
	replyCh, err := client.openDataConnection(addr, dataAction)
	if err != nil {
		return err
	}

	buf := make([]byte, 0, 8196)
	n := 0
	switch dataAction {
	case 'r':
		for tmp := range replyCh {
			buf = append(buf, tmp...)
			n += len(tmp)
			client.ioCh <- fmt.Sprintf("\rDownloaded: %d bytes.", n)
		}
		client.ioCh <- "\n"
	case 'w':
		bytes, err := ioutil.ReadFile(arg)
		if err != nil {
			return err
		}
		replyCh <- bytes
		close(replyCh)
	}

	switch cmd {
	case ftp_cmd.LIST:
		client.ioCh <- fmt.Sprintf(string(buf))
	case ftp_cmd.RETR:
		pathComponents := strings.Split(arg, "/")
		name := pathComponents[len(pathComponents)-1]
		err = ioutil.WriteFile(client.outDir+name, buf, 0644)
	case ftp_cmd.STOR:
		client.ioCh <- fmt.Sprintf("File %s saved to server.\n", arg)
	default:
		return errors.New("Unknown command")
	}
	wg.Done()
	return nil
}

func (client *FtpClient) openDataConnection(addr string, mode byte) (chan []byte, error) {
	conn, err := client.getDataConnection(addr)
	if err != nil {
		return nil, err
	}
	switch mode {
	case 'r':
		return readDataAsync(conn), nil
	case 'w':
		return writeDataAsync(conn), nil
	default:
		return nil, fmt.Errorf("Invalid mode %c", mode)
	}
}

func (client *FtpClient) readCtrlConn() (int, string, error) {
	if !client.ctrlConnScanner.Scan() {
		return 0, "", errors.New("Server connection closed")
	}
	line := strings.SplitN(client.ctrlConnScanner.Text(), " ", 2)
	status, err := strconv.Atoi(line[0])
	if err != nil {
		return 0, "", err
	}
	log.Printf("Reading from srv %d %s.\n", status, line[1])
	return status, line[1], nil
}

func readDataAsync(r net.Conn) chan []byte {
	ch := make(chan []byte)
	go func() {
		for {
			tmp := make([]byte, 256)
			n, err := r.Read(tmp)
			if err != nil {
				if err != io.EOF {
					log.Println("read error:", err)
				}
				close(ch)
				r.Close()
				return
			}
			ch <- tmp[:n]
		}
	}()
	return ch
}

func writeDataAsync(r net.Conn) chan []byte {
	ch := make(chan []byte)
	go func() {
		for {
			for d := range ch {
				_, err := r.Write(d)
				if err != nil {
					log.Println(err.Error())
					return
				}
			}
			r.Close()
		}
	}()
	return ch
}

func (client *FtpClient) readReply() (int, string, error) {
	var status int
	var reply string
	var err error
	for {
		status, reply, err = client.readCtrlConn()
		if err != nil {
			return 0, "", err
		}
		if status != 150 {
			client.connectionMode = ftp_cmd.NOT_SET
			break
		}
	}
	return status, reply, err
}

func (client *FtpClient) getDataConnection(addr string) (net.Conn, error) {
	var conn net.Conn
	var err error
	switch client.connectionMode {
	case ftp_cmd.ACTIVE:
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		conn, err = ln.Accept()
	case ftp_cmd.PASSIVE:
		conn, err = net.Dial("tcp", addr)
	default:
		return nil, errors.New("Unknown connection mode")
	}
	return conn, err
}

func (client *FtpClient) send(cmd ftp_cmd.CmdType, args string) (int, error) {
	return fmt.Fprintf(client.ctrlConn, string(cmd)+" "+args+"\r\n")
}

func unexpectedStatusError(received, expected int) error {
	return fmt.Errorf("Expected status code %d when sending user to server but received status %d", expected, received)
}

func getDataAction(cmd ftp_cmd.CmdType) (byte, error) {
	switch cmd {
	case ftp_cmd.RETR, ftp_cmd.LIST:
		return 'r', nil
	case ftp_cmd.STOR:
		return 'w', nil
	default:
		return ' ', errors.New("Invalid command")
	}
}
