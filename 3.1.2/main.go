package main

import (
	"encoding/json"
	"log"
	"net/http"

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
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *server) createPost(w http.ResponseWriter, r *http.Request) {
	var p db.Post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Fatal(err)
	}
	s.dbase.CreatePost(p)
	json.NewEncoder(w).Encode(p)
}

func (s *server) getPosts(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.dbase.GetPosts())
}
