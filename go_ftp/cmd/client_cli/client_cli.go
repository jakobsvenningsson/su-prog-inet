package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_client"
)

func main() {
	// Parse command line flags
	var (
		pUser         = flag.String("u", "", "FTP server user name")
		pPw           = flag.String("pw", "", "FTP server password")
		pIt           = flag.Bool("it", false, "Interactive mode")
		pIsLogEnabled = flag.Bool("log", false, "Enable logging")
		pDataIP       = flag.String("data-ip", "", "Data socket addr")
		pOutDir       = flag.String("out", "./", "The folder which downloads will be saved.")
	)
	flag.Parse()
	user, pw, it, isLogEnabled, dataIP, outDir := *pUser, *pPw, *pIt, *pIsLogEnabled, *pDataIP, *pOutDir

	if len(user) == 0 || len(pw) == 0 {
		fmt.Printf("Please specify login credentials using the -u and -pw flags.\n")
		os.Exit(1)
	}

	if !isLogEnabled {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("Starting FTP client, interactive mode %t.", it)
	srvAddr, cmds := parseCommandlineArguments(it)
	client, conn, err := initFtpClient(srvAddr, dataIP, outDir, cmds, it)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Read server welcome message
	status, welcomeMsg, err := client.ReadWelcomeMessage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d %s\n", status, welcomeMsg)

	// Try to authenticate
	if err := client.Authenticate(user, pw); err != nil {
		log.Fatal(err)
	}
	log.Printf("Authentication successful.\n")

	log.Fatal(client.ProcessCommands())
}

// Helpers

func inputCommandReader(it bool, cmds string) io.Reader {
	if it {
		return os.Stdin
	}
	return strings.NewReader(cmds)
}

func parseCommandlineArguments(it bool) (string, string) {
	args := flag.Args()
	srvAddr := args[0]
	if it {
		return srvAddr, ""
	}
	if len(args) < 2 {
		fmt.Printf("No command supplied.\n")
		os.Exit(1)
	}
	cmds := strings.Replace(args[1], ";", "\n", -1)
	return srvAddr, cmds
}

func initFtpClient(srvAddr, dataIP, outDir, cmds string, it bool) (*ftp_client.FtpClient, net.Conn, error) {
	usrIn := inputCommandReader(it, cmds)
	ctrlConn, err := net.Dial("tcp", srvAddr)
	if err != nil {
		return nil, nil, err
	}
	client, err := ftp_client.New(usrIn, ctrlConn, dataIP, outDir)
	if err != nil {
		return nil, nil, err
	}
	return client, ctrlConn, nil
}
