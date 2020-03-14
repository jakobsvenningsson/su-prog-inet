package main

import (
	"flag"
	"log"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_server"
)

var (
	pRoot = flag.String("root", "/tmp", "Root directory of FTP server.")
	pPort = flag.String("port", "10000", "Control connection port.")
	pIP   = flag.String("ip", "", "Control connection addr.")
)

func main() {
	flag.Parse()
	root, port, ip := *pRoot, *pPort, *pIP
	log.Printf("Starting FTP server on port: %s, with root: %s.\n", ip+":"+port, root)
	ftpserver := ftp_server.New(root, ip, port)
	log.Fatal(ftpserver.Start())
}
