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
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_ip"
)

const IOSync = "IO.SYNC"

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
	var wg sync.WaitGroup
	go client.startIOChannel(&wg)
	for {
		cmd, err := client.usrIn.NextCommand()
		if err != nil {
			log.Printf("%s.\n", err.Error())
			continue
		}
		log.Printf("Processing cmd: %s, arg: %s.\n", cmd.Type, cmd.Arg)
		status, reply, err := client.processCommand(cmd, &wg)
		if err != nil {
			log.Printf("%s.\n", err.Error())
			continue
		}
		client.ioCh <- fmt.Sprintf("\r%d %s\n", status, reply)
	}
	wg.Wait()
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
	return client.read()
}

// Private Methods

func (client *FtpClient) processCommand(command *ftp_cmd.Cmd, wg *sync.WaitGroup) (int, string, error) {
	cmd, arg := command.Type, command.Arg
	switch cmd {
	case ftp_cmd.LIST, ftp_cmd.RETR:
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
	case ftp_cmd.PASS, ftp_cmd.USER, ftp_cmd.CWD, ftp_cmd.PWD, ftp_cmd.PASV, ftp_cmd.QUIT:
	default:
		return 0, "", errors.New("Command not implemented")
	}

	if _, err := client.write(cmd, arg); err != nil {
		return 0, "", err
	}

	status, reply, err := client.handleReply()
	if status == 226 {
		status, reply, err = client.handleReply()
	}
	if err != nil {
		return 0, "", err
	}
	return status, reply, nil
}

func (client *FtpClient) handleReply() (int, string, error) {
	status, reply, err := client.read()
	if err != nil {
		return 0, "", err
	}

	switch status {
	case 227:
		addr, err := ftp_ip.Decode(reply)
		if err != nil {
			return 0, "", err
		}
		client.dataConnAddr = addr
		client.connectionMode = ftp_cmd.PASSIVE
	case 221:
		return status, reply, errors.New("Time to quit")
	case 150:
		client.connectionMode = ftp_cmd.NOT_SET
	default:
	}
	return status, reply, nil
}

func (client *FtpClient) startDataConnection(cmd ftp_cmd.CmdType, args, addr string, wg *sync.WaitGroup) error {
	wg.Add(2)
	log.Printf("Starting data channel on addr %s.\n", addr)
	replyCh, err := client.openDataConnection(addr)
	if err != nil {
		return err
	}

	buf := make([]byte, 0, 8196)
	n := 0
	for tmp := range replyCh {
		buf = append(buf, tmp...)
		n += len(tmp)
		client.ioCh <- fmt.Sprintf("\rDownloaded: %d bytes.", n)
	}
	client.ioCh <- "\n"

	switch cmd {
	case ftp_cmd.LIST:

		client.ioCh <- fmt.Sprintf(string(buf))
	case ftp_cmd.RETR:
		pathComponents := strings.Split(args, "/")
		name := pathComponents[len(pathComponents)-1]
		err = ioutil.WriteFile(client.outDir+name, buf, 0644)
	default:
		log.Fatal("unknown command")
	}
	client.ioCh <- IOSync
	wg.Done()
	return nil
}

func (client *FtpClient) openDataConnection(addr string) (chan []byte, error) {
	var conn net.Conn
	var err error
	switch client.connectionMode {
	case ftp_cmd.ACTIVE:
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		conn, err = ln.Accept()
		if err != nil {
			return nil, err
		}
	case ftp_cmd.PASSIVE:
		conn, err = net.Dial("tcp", addr)
	default:
		return nil, errors.New("Unknown connection model")
	}
	if err != nil {
		return nil, err
	}
	return readDataAsync(conn), nil
}

func (client *FtpClient) read() (int, string, error) {
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

func (client *FtpClient) startIOChannel(wg *sync.WaitGroup) {
	for s := range client.ioCh {
		if s == IOSync {
			wg.Done()
			continue
		}
		fmt.Printf(s)
	}
}

func (client *FtpClient) nextCommand() (ftp_cmd.CmdType, string, error) {
	cmd, err := client.usrIn.NextCommand()
	return cmd.Type, cmd.Arg, err
}

func (client *FtpClient) write(cmd ftp_cmd.CmdType, args string) (int, error) {
	return fmt.Fprintf(client.ctrlConn, string(cmd)+" "+args+"\r\n")
}

func unexpectedStatusError(received, expected int) error {
	return fmt.Errorf("Expected status code %d when sending user to server but received status %d", expected, received)
}
