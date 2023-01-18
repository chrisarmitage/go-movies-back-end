package models

import "time"

type User struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

