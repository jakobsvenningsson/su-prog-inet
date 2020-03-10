package ftp_ip

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"regexp"
)

func Decode(addrString string) string {
	log.Printf("PARSING ADDR %s\n", addrString)
	re := regexp.MustCompile(`(\d+,?){6}`)
	components := strings.Split(re.FindString(addrString), ",")
	host := strings.Join(components[:4], ".")

	portPart1, err := strconv.Atoi(components[4])
	if err != nil {
		log.Fatal(err)
	}
	portPart2, err := strconv.Atoi(components[5])
	if err != nil {
		log.Fatal(err)
	}

	port := portPart1*256 + portPart2

	return fmt.Sprintf("%s:%d", host, port)
}

func Encode(addrString string) string {
	c := strings.Split(addrString, ":")
	host := strings.Split(c[0], ".")
	log.Println(host)
	port, err := strconv.Atoi(c[1])
	if err != nil {
		log.Fatal(err)
	}
	p1 := port % 256
	p2 := (port - p1) / 256

	return fmt.Sprintf("%s,%s,%s,%s,%d,%d", host[0], host[1], host[2], host[3], p2, p1)
}
