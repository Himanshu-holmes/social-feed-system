package models

import "time"

type User struct {
	ID       string
	Username string
}

type Post struct {
	ID        string
	Content   string
	Timestamp time.Time
	AuthorID  string
}