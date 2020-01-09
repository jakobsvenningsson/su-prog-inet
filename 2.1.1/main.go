package main 

import (
    "2.1.1/client"
    "fmt"
    "os"
)

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

func main() {

    // Read chat-server host and port provided by user, or fallback to default values.
    host, port := getHostAndPort()

    // Create a new chat-client.
    client := client.New(host, port)

    // Try to connect to chat-server.
    err := client.Connect()
    if err != nil {
        fmt.Printf("Connection failed, exiting...")
        return 
    }
    fmt.Printf("Connection successful.\n")

    // Start listening for user input, and send input input to chat-server. 
    client.StartClientSendLoop()

    // Start listening for chat messages from other client.
    // This function blocks.
    client.ListenForServerMessages()
}
