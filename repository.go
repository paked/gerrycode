package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

// Repository is the representation of a git project on Rr
type Repository struct {
	ID   bson.ObjectId `bson:"_id" json:"_id"`
	Host string        `bson:"host" json:"host"`
	User string        `bson:"user" json:"user"`
	Name string        `bson:"name" json:"name"`
}

// NewRepository creates a new Repository link.
// 		POST /api/repo/{host}/{user}/{name}?access_token=xxx
func NewRepository(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]
	e := json.NewEncoder(w)

	c := server.Collection("repositories")

	var re Repository
	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re != (Repository{}) {
		e.Encode(Response{Message: "That repo already exists", Status: NewFailedStatus()})
		return
	}

	re = Repository{ID: bson.NewObjectId(), Host: host, User: user, Name: name}
	if err := c.Insert(re); err != nil {
		e.Encode(Response{Message: "Could not insert that repository", Status: NewServerErrorStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "We created that new repository!", Status: NewOKStatus(), Data: re})
}

// GetRepository retrieves a Repository.
// 		GET /api/repo/{host}/{user}/{name}
func GetRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]
	e := json.NewEncoder(w)

	c := server.Collection("repositories")

	var re Repository
	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re == (Repository{}) {
		e.Encode(Response{Message: "That repository does not exist", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is your repo", Status: NewOKStatus(), Data: re})
}
