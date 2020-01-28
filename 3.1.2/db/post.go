package db

type Post struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Homepage string `json:"homepage"`
	Comment  string `json:"comment"`
}
