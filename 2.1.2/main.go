package main

import (
    "2.1.2/server"
    "log"
)

func main() {
    server := server.New("2000")
    log.Fatal(server.Listen())
}
