package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"../ftpclient"
)

var (
	user         = flag.String("u", "", "FTP server user name")
	pw           = flag.String("pw", "", "FTP server password")
	it           = flag.Bool("it", false, "Interactive mode")
	isLogEnabled = flag.Bool("log", false, "Enable logging")
	dataAddr = flag.String("data-addr", "", "Data socket addr")

)

func main() {
	url, usr, pw, it, isLogEnabled, dataAddr, cmds := parseCmdArgs()
	if !isLogEnabled {
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("Starting FTP client, interactive mode %t.", it)

	input := inputCommandReader(it, cmds)
	client, err := ftpclient.New(url, input, dataAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Conn().Close()

	// Read server welcome message
	log.Printf("Reading welcone message.\n")
	status, welcomeMsg, err := client.ReadWelcomeMessage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d %s.\n", status, welcomeMsg)

	// Try to authenticate
	if err := client.Authenticate(usr, pw); err != nil {
		log.Fatal(err)
	}
	log.Printf("Authentication successfull.\n")

	client.ProcessCommands()
}

// Helpers

func parseCmdArgs() (string, string, string, bool, bool, string, string) {
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

	return ftpServer, *user, *pw, *it, *isLogEnabled, *dataAddr, cmds
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
