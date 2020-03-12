package ftp_ip

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func Decode(addr string) (string, error) {
	re := regexp.MustCompile(`(\d+,){5}\d+`)
	str := re.FindString(addr)
	if str == "" {
		return "", errors.New("Invalid addr format")
	}
	components := strings.Split(str, ",")
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

	return fmt.Sprintf("%s:%d", host, port), nil
}

func Encode(addr string) (string, error) {
	re := regexp.MustCompile(`\d+.\d+.\d+.\d+:\d+`)
	matched := re.MatchString(addr)
	if !matched {
		return "", errors.New("Invalid addr format")
	}
	c := strings.Split(addr, ":")
	host := strings.Split(c[0], ".")
	log.Println(host)
	port, err := strconv.Atoi(c[1])
	if err != nil {
		log.Fatal(err)
	}
	p1 := port % 256
	p2 := (port - p1) / 256

	return fmt.Sprintf("%s,%s,%s,%s,%d,%d", host[0], host[1], host[2], host[3], p2, p1), nil
}
