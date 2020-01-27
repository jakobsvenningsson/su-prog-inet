package main

import (
	"log"
	"os"
	"strings"

	server "2.2.1/drawserver"
)

func main() {
	port, peers := os.Args[1], os.Args[2]
	server := server.New(port, strings.Split(peers, ","))
	log.Fatal(server.Listen())
}
