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

var (
	user         = flag.String("u", "", "FTP server user name")
	pw           = flag.String("pw", "", "FTP server password")
	it           = flag.Bool("it", false, "Interactive mode")
	isLogEnabled = flag.Bool("log", false, "Enable logging")
	dataAddr     = flag.String("data-addr", "", "Data socket addr")
	outDir       = flag.String("out", "./", "The folder which downloads will be saved.")
)

func main() {
	ctrlAddr, usr, pw, it, isLogEnabled, dataAddr, outDir, cmds := parseCmdArgs()
	if !isLogEnabled {
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("Starting FTP client, interactive mode %t.", it)

	usrIn := inputCommandReader(it, cmds)
	ctrlConn, err := net.Dial("tcp", ctrlAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ctrlConn.Close()
	client, err := ftp_client.New(usrIn, ctrlConn, dataAddr, outDir)
	if err != nil {
		log.Fatal(err)
	}

	// Read server welcome message
	status, welcomeMsg, err := client.ReadWelcomeMessage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d %s\n", status, welcomeMsg)

	// Try to authenticate
	if err := client.Authenticate(usr, pw); err != nil {
		log.Fatal(err)
	}
	log.Printf("Authentication successfull.\n")

	log.Fatal(client.ProcessCommands())
}

// Helpers

func parseCmdArgs() (string, string, string, bool, bool, string, string, string) {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Please specify a FTP server.\n")
		os.Exit(1)
	}
	ftpServer := args[0]

	if len(*user) == 0 || len(*pw) == 0 {
		fmt.Printf("Please specify login credentials using the -u and -pw flags.\n")
		os.Exit(1)
	}
	var cmds string
	if !*it {
		if len(args) < 2 {
			fmt.Printf("No command.\n")
			os.Exit(1)
		}
		cmds = args[1]
	}

	return ftpServer, *user, *pw, *it, *isLogEnabled, *dataAddr, *outDir, cmds
}

func inputCommandReader(it bool, cmds string) io.Reader {
	var input io.Reader
	if it {
		input = os.Stdin
	} else {
		input = strings.NewReader(cmds)
	}
	return input
}
