package ftpclient

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

	"../ftpcommand"
	"../parser"
	"../ftp_ip"

)

type FtpClient struct {
	conn      net.Conn
	srvIn     *bufio.Scanner
	usrIn     *parser.Parser 
	connectionMode      ftpcommand.MODE
	dataConnAddr  string
}

// Public Methods

func New(url string, input io.Reader, dataConnAddr string) (*FtpClient, error) {
	conn, err := net.Dial("tcp", url+":21")
	if err != nil {
		return nil, err
	}
	usrIn := parser.New(input)
	srvIn := bufio.NewScanner(conn)
	return &FtpClient{conn, srvIn, usrIn, ftpcommand.ACTIVE, dataConnAddr}, nil
}

func (client *FtpClient) ProcessCommands() {
	var wg sync.WaitGroup
	for {
		token, err := client.usrIn.NextToken()
		if err != nil {
			break
		}
		log.Printf("Processing token: %s, arg: %s.\n", token.Type, token.Arg)
		status, reply, err := client.processToken(token, &wg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d %s\n", status, reply)
	}
	wg.Wait()
}

func (client *FtpClient) Authenticate(user, pw string) error {
	// 1. Send user using the "USER :user" FTP command
	status, _, err := client.processToken(&parser.Token{ftpcommand.USER, user}, nil)
	if err != nil {
		return err
	}
	if status != 331 {
		return unexpectedStatusError(status, 331)
	}
	// 2. Send password using the "PASS :password" FTP command
	status, _, err = client.processToken(&parser.Token{ftpcommand.PASS, pw}, nil)
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

func (client *FtpClient) Conn() net.Conn {
	return client.conn
}

// Private Methods

func (client *FtpClient) processToken(token *parser.Token, wg *sync.WaitGroup) (int, string, error) {
	fmt.Printf("Processing %s %s.\n", token.Type, token.Arg)

	cmd, arg := token.Type, token.Arg
	if (cmd.IsDataCMD() && client.connectionMode == ftpcommand.PASSIVE) || cmd == ftpcommand.PORT {
		go client.startDataConnection(cmd, arg, client.dataConnAddr, wg)
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
		client.dataConnAddr = ftp_ip.Decode(reply)
		client.connectionMode = ftpcommand.PASSIVE
	case 221:
		return status, reply, errors.New("Time to quit")
	default:
		client.connectionMode = ftpcommand.ACTIVE
	}
	return status, reply, nil
}

func (client *FtpClient) startDataConnection(cmd ftpcommand.CMD, args, addr string, wg *sync.WaitGroup) error {
	wg.Add(1)
	log.Printf("Starting data channel on addr %s.\n", addr)
	replyCh, err := client.openDataConnection(addr)
	if err != nil {
		return err
	}
	buf := make([]byte, 0, 8196)
	n := 0
	log.Printf("Trying to read...")
	for tmp := range replyCh {
		buf = append(buf, tmp...)
		n += len(tmp)
		fmt.Printf("\rDownloaded: %d bytes.", n)
	}
	fmt.Println()

	switch cmd {
	case ftpcommand.LIST:
		fmt.Printf(string(buf))
	case ftpcommand.RETR:
		pathComponents := strings.Split(args, "/")
		name := pathComponents[len(pathComponents)-1]
		err = ioutil.WriteFile("./"+name, buf, 0644)
	}

	wg.Done()

	return nil
}

func (client *FtpClient) openDataConnection(addr string) (chan []byte, error) {
	var conn net.Conn
	var err error
	switch client.connectionMode {
	case ftpcommand.ACTIVE:
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		conn, err = ln.Accept()
	case ftpcommand.PASSIVE:
		conn, err = net.Dial("tcp", addr)
	default:
		return nil, errors.New("Unknown connection mode.")
	}
	if err != nil {
		return nil, err
	}
	return readDataAsync(conn), nil
}

func (client *FtpClient) read() (int, string, error) {
	if !client.srvIn.Scan() {
		return 0, "", errors.New("Could not read server reply")
	}
	line := strings.SplitN(client.srvIn.Text(), " ", 2)
	status, err := strconv.Atoi(line[0])
	if err != nil {
		return 0, "", err
	}
	return status, line[1], nil
}

func readDataAsync(r net.Conn) chan []byte {
	ch := make(chan []byte)
	go func() {
		tmp := make([]byte, 256)
		for {
			n, err := r.Read(tmp)
			if err != nil {
				if err != io.EOF {
					fmt.Println("read error:", err)
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

func (client *FtpClient) nextCommand() (ftpcommand.CMD, string, error) {
	token, err := client.usrIn.NextToken()
	return token.Type, token.Arg, err
}

func (client *FtpClient) write(cmd ftpcommand.CMD, args string) (int, error) {
	return fmt.Fprintf(client.conn, string(cmd)+" "+args+"\r\n")
}

func unexpectedStatusError(received, expected int) error {
	return fmt.Errorf("Expected status code %d when sending user to server but received status %d", expected, received)
}