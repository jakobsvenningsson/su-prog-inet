// Package Server implements a simple chat-webserver.
// The webserver is meant to be used with the client impelemented in assigment 2.1.1.
package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Server represents a chat server.
type Server struct {
	port    string
	mux     sync.Mutex
	clients []net.Conn
}

// Public Methods

// New initialized and return the address of a new struct Server.
func New(port string) *Server {
	return &Server{
		port:    port,
		clients: []net.Conn{},
	}
}

// Listen makes the server start listening for new incomming connections.
// Each client is handled in a seperate thread.
func (s *Server) Listen() error {
	socket, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return err
	}
	for {
		conn, err := socket.Accept()
		if err != nil {
			fmt.Printf("Error accepting new client.\n")
			continue
		}
		s.addClient(conn)
		go s.handleNewClient(conn)
	}
}

// Private Methods

func (s *Server) handleNewClient(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Connection closed.\n")
			s.removeClient(conn)
			return
		}
		fmt.Printf("Message: %s, from %s.\n", strings.TrimSuffix(message, "\n"), conn.RemoteAddr())
		// Broadcast new message to all known clients
		for _, client := range s.clients {
			fmt.Fprintf(client, message)
		}
	}
}

// addClient appends a new client to the list of connected clients
func (s *Server) addClient(conn net.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.clients = append(s.clients, conn)
}

// removeClient removes a client from the list of connected clients
func (s *Server) removeClient(conn net.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for i, client := range s.clients {
		if client == conn {
			// Remove client from list
			s.clients[i], s.clients[len(s.clients)-1] = s.clients[len(s.clients)-1], s.clients[i]
			fmt.Printf("Removing client %s.\n", conn.RemoteAddr().String())
			break
		}
	}
}
