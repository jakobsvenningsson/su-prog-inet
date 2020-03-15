// Package client implements a command-line based chat-client.
// The chat-client is meant to be used with the chat-server implemented in assigment 2.1.2.
package client

import (
	"fmt"
	"net"
)

// Client represents a server connection.
type Client struct {
	host string
	port string
	conn net.Conn
}

// Public Methods

// New initializes and returns the address of a new struct Client.
func New(host, port string) *Client {
	return &Client{host, port, nil}
}

// Connect attempts to establish a connection to the host and port specified in struct Client's host and port fields.
func (c *Client) Connect() error {
	fmt.Printf("Trying to connect to %s.\n", c.srvAddr())
	conn, err := net.Dial("tcp", c.srvAddr())
	c.conn = conn
	return err
}

// Conn returns the net.Conn object of the client
func (c *Client) Conn() net.Conn {
	return c.conn
}

// Private Methods

func (c *Client) srvAddr() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}
