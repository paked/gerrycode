package main

import (
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
)

// Repository is the representation of a git project on Rr
type Repository struct {
	ID   bson.ObjectId `bson:"_id" json:"_id"`
	Host string        `bson:"host" json:"host"`
	User string        `bson:"user" json:"user"`
	Name string        `bson:"name" json:"name"`
}

func (rep Repository) BID() bson.ObjectId {
	return rep.ID
}

func (rep Repository) C() string {
	return "repositories"
}

// NewRepository creates a new Repository link.
// 		POST /api/repo/{host}/{user}/{name}?access_token=xxx
func NewRepository(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]
	e := json.NewEncoder(w)

	var rep Repository
	if err := models.Restore(&rep, bson.M{"host": host, "user": user, "name": name}); err == nil {
		e.Encode(Response{Message: "That repo already exists", Status: NewFailedStatus()})
		return
	}

	rep = Repository{ID: bson.NewObjectId(), Host: host, User: user, Name: name}
	if err := models.Persist(rep); err != nil {
		e.Encode(Response{Message: "Could not insert that repository", Status: NewServerErrorStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "We created that new repository!", Status: NewOKStatus(), Data: rep})
}
