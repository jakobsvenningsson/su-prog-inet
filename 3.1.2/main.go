package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"3.1.2/db"
)

type database interface {
	CreatePost(p db.Post)
	GetPosts() []db.Post
}

type server struct {
	dbase database
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify a port.")
	}
	port := os.Args[1]
	s := server{dbase: db.New()}
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			s.createPost(w, r)
		case "GET":
			s.getPosts(w, r)
		}
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// createPost creates a post in the database using the data supplied from an HTTP POST reqest.
func (s *server) createPost(w http.ResponseWriter, r *http.Request) {
	var p db.Post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Fatal(err)
	}
	s.dbase.CreatePost(p)
	json.NewEncoder(w).Encode(p)
}

// getPost returns all database posts as JSON encoded data
func (s *server) getPosts(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.dbase.GetPosts())
}
