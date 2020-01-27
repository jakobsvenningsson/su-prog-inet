package main

import (
	"log"
	"os"

	server "2.2.1/drawserver"
)

func main() {
	port, peerHost, peerPort := os.Args[1], os.Args[2], os.Args[3]
	server := server.New(port, []server.Peer{server.NewPeer(peerHost, peerPort)})
	log.Fatal(server.Listen())
}
