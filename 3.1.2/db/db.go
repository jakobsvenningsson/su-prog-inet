package db

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	posts []Post
	conn  *sql.DB
}

func New() *DB {
	const (
		host     = "atlas.dsv.su.se"
		port     = 3306
		user     = "usr_20849871"
		password = "849871"
		dbname   = "db_20849871"
	)

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbname)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS Posts ( id int NOT NULL AUTO_INCREMENT, " +
			"name varchar(128), email varchar(128), homepage varchar(128), comment varchar(128), PRIMARY KEY (id));")
	if err != nil {
		log.Fatal(err)
	}

	return &DB{posts: []Post{}, conn: db}
}

func (db *DB) CreatePost(p Post) {

	r, _ := regexp.Compile("<([^>]+)>")
	p.Name = r.ReplaceAllString(p.Name, "censur")
	p.Email = r.ReplaceAllString(p.Email, "censur")
	p.Homepage = r.ReplaceAllString(p.Homepage, "censur")
	p.Comment = r.ReplaceAllString(p.Comment, "censur")

	queryString := fmt.Sprintf("INSERT INTO Posts (name, email, homepage, comment) VALUES ('%s', '%s', '%s', '%s')", p.Name, p.Email, p.Homepage, p.Comment)
	insert, err := db.conn.Query(queryString)
	if err != nil {
		log.Fatal(err)
	}
	defer insert.Close()
}

func (db *DB) GetPosts() []Post {
	results, err := db.conn.Query("SELECT * FROM Posts")
	if err != nil {
		log.Fatal(err)
	}
	posts := []Post{}
	for results.Next() {
		var post Post
		err = results.Scan(&post.ID, &post.Name, &post.Email, &post.Homepage, &post.Comment)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, post)
	}
	return posts
}
