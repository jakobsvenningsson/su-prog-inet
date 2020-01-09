// Package Server implements a simple chat-webserver. 
// The webserver is meant to be used with the client impelemented in assigment 2.1.1.
package server

import (
    "net"
    "fmt"
    "sync"
    "bufio"
)

// Server represents a webserver.
type Server struct {
    port string 
    mux sync.Mutex
    clients []net.Conn
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
            continue;
        }
        s.addClient(conn)
        s.handleNewClient(conn)
    }
    return err
}

// New initialized and return the address of a new struct Server.
func New(port string) *Server {
    return &Server{
        port: port, 
        clients: []net.Conn{},
    }
}

func (s *Server) handleNewClient(conn net.Conn) {
    reader := bufio.NewReader(conn)
    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("Connection closed.\n")
            s.removeClient(conn)
            return
        }
        for _, client := range s.clients {
            fmt.Fprintf(client, message)
        }
    }
}

func (s *Server) addClient(conn net.Conn) {
    s.mux.Lock()
    defer s.mux.Unlock()
    s.clients = append(s.clients, conn)
}

func (s *Server) removeClient(conn net.Conn) {
    s.mux.Lock()
    defer s.mux.Unlock()
    for i, client := range s.clients {
        if client == conn {
            s.clients[i], s.clients[len(s.clients) - 1] = s.clients[len(s.clients) - 1], s.clients[i] 
            fmt.Printf("Removing client %s.\n", conn.RemoteAddr().String())
            break
        }
    }
}
