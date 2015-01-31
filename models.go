package main

import (
	"gopkg.in/mgo.v2/bson"
)

// User is someone who has registered on the site.
type User struct {
	ID           bson.ObjectId `bsttqon:"_id" json:"_id"`
	Username     string        `bson:"username" json:"username"`
	PasswordHash string        `bson:"password_hash" json:"-"`
	Email        string        `bson:"email" json:"email"`
	PasswordSalt string        `bson:"password_salt" json:"-"`
}

// A Review is created by a User to express their feelings of a particular Repository
type Review struct {
	ID         bson.ObjectId `bson:"_id" json:"_id"`
	From       bson.ObjectId `bson:"from" json:"from"`
	Repository bson.ObjectId `bson:"repository" json:"repository"`
	Content    string        `bson:"content" json:"content"`
	Rating     int           `bson:"rating" json:"rating"`
}

// Repository is the representation of a git project on Rr
type Repository struct {
	ID   bson.ObjectId `bson:"_id" json:"_id"`
	Host string        `bson:"host" json:"host"`
	User string        `bson:"user" json:"user"`
	Name string        `bson:"name" json:"name"`
}

// Token is a container used to send a User their access_token
type Token struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
