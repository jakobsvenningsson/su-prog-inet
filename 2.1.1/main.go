package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"2.1.1/client"
)

func main() {

	// Read chat-server host and port provided by user, or fallback to default values.
	host, port := getHostAndPort()

	// Create a new chat-client.
	c := client.New(host, port)

	// Try to connect to chat-server.
	err := c.Connect()
	if err != nil {
		fmt.Printf("Connection failed, exiting...")
		return
	}
	fmt.Printf("Connection successful.\n")

	// Start listening for user input, and send input input to chat-server.
	go startReadWriteLoop(os.Stdin, c.Conn(), func(w io.Writer, s string) {
		fmt.Fprintf(w, s)
	})

	// Start listening for chat messages from other client.
	go startReadWriteLoop(c.Conn(), os.Stdout, func(w io.Writer, s string) {
		fmt.Fprintf(w, "Message from server: %s", s)
	})

	// Block indefinitly
	select {}
}

func startReadWriteLoop(in io.Reader, out io.Writer, action func(io.Writer, string)) {
	reader := bufio.NewReader(in)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error %s.\n", err)
			return
		}
		action(out, s)
	}
}

func getHostAndPort() (string, string) {
	host, port := "127.0.0.1", "2000"
	switch len(os.Args) {
	case 2:
		host = os.Args[1]
	case 3:
		host, port = os.Args[1], os.Args[2]
	default:
	}
	return host, port
}
