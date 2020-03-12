package main

import (
	"flag"
	"log"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_server"
)

var (
	root = flag.String("root", "/tmp", "Root directory of FTP server.")
	port = flag.String("port", "10000", "Control connection port.")
	ip   = flag.String("ip", "127.0.0.1", "Control connection addr.")
)

func main() {
	root, port, ip := parseCmdArgs()
	log.Printf("Starting FTP server on port %s with root %s.\n", root, ip+":"+port)
	ftpserver := ftp_server.New(root, ip, port)
	log.Fatal(ftpserver.Start())
}

func parseCmdArgs() (string, string, string) {
	flag.Parse()
	return *root, *port, *ip
}
