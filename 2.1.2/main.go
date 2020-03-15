package main

import (
	"fmt"
	"log"
	"os"

	"2.1.2/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify a port.")
		os.Exit(1)
	}
	port := os.Args[1]
	s := server.New(port)
	log.Fatal(s.Listen())
}
