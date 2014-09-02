package main

import "time"

type Asset struct {
	ID          string           `json:"id"`
	CreatedTime string           `json:"created_time"`
	Images      map[string]Image `json:"images"`
	Created     time.Time        `json:"-"`
	User        User             `json:"user"`
}

type User struct {
	Username string `json:"username"`
}

type Image struct {
	URL string `json:"url"`
}
