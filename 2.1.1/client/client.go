// Package client implements a command-line based chat-client.
// The chat-client is meant to be used with the chat-server implemented in assigment 2.1.2. 
package client

import (
    "fmt"
    "os"
    "net"
    "bufio"
)

// Client represents a server connection.
type Client struct {
    host string
    port string 
    conn net.Conn
}

// Connect attempts to establish a connection to the host and port specified in struct Client's host and port fields.
func (c *Client) Connect() error {
    fmt.Printf("Trying to connect to %s.\n", c.srvAddr())
    conn, err := net.Dial("tcp", c.srvAddr())
    c.conn = conn
    return err
}

// ListenForServerMessages starts a new goroutine (thread) that will listen for incoming server messages.
// The incoming messages will be printed to stdout.
func (c *Client) ListenForServerMessages() {
    reader := bufio.NewReader(c.conn)
    go func() {
        for {
            message, err := reader.ReadString('\n')
            if err != nil {
                fmt.Printf("Error reading message.\n")
            }
            fmt.Printf("Message from server: %s", message)
        }
    }()
}

// StartClientSendLoop listens for input on stdin and writes user input to the chat-server.
func (c *Client) StartClientSendLoop() {
    reader := bufio.NewReader(os.Stdin)
    for {
        text, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("Error reading user input data.\n")
            continue
        }
        fmt.Fprintf(c.conn, text)
    }
}

// New initializes and returns the address of a new struct Client.
func New(host, port string) *Client {
    return &Client{host, port, nil}
}

func (c *Client) srvAddr() string {
    return fmt.Sprintf("%s:%s", c.host, c.port)
}
