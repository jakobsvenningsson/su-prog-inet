package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

type emailContext struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Message  string `json:"message"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify a port.")
		os.Exit(1)
	}
	port := os.Args[1]

	http.Handle("/", http.FileServer(http.Dir("static")))
	// Post Email Endpoint
	http.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
		var eCtx emailContext
		if err := json.NewDecoder(r.Body).Decode(&eCtx); err != nil {
			log.Fatal(err)
		}
		auth := smtp.PlainAuth("", eCtx.User, eCtx.Password, eCtx.Server)
		to := []string{eCtx.To}
		msg := []byte("To: " + eCtx.To + "\r\n" +
			"Subject: " + eCtx.Subject + "\r\n" +
			"\r\n" +
			eCtx.Message + "\r\n")
		fmt.Println(string(msg))

		if err := smtp.SendMail(eCtx.Server+":"+eCtx.Port, auth, eCtx.From, to, msg); err != nil {
			log.Fatal(err)
		}
	})
	// Serve HTTP
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
