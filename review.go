package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

// A Review is created by a User to express their feelings of a particular Repository
type Review struct {
	ID         bson.ObjectId `bson:"_id" json:"_id"`
	From       bson.ObjectId `bson:"from" json:"from"`
	Repository bson.ObjectId `bson:"repository" json:"repository"`
	Content    string        `bson:"content" json:"content"`
	Rating     int           `bson:"rating" json:"rating"`
}

func (rev Review) C() string {
	return "reviews"
}

func (rev Review) BID() bson.ObjectId {
	return rev.ID
}

// NewReviewHandler creates a new Review on a Repository.
// 		POST /api/repo/{repository}/review?text=This+sucks&rating=2&access_token=xxx
func NewReviewHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name, review := vars["host"], vars["user"], vars["name"], r.FormValue("review")
	e := json.NewEncoder(w)

	if review == "" {
		e.Encode(Response{Message: "Please say something in your review :)", Status: NewFailedStatus()})
		return
	}

	var rep Repository
	if err := models.Restore(&rep, bson.M{"host": host, "user": user, "name": name}); err != nil {
		e.Encode(Response{Message: "A repo with that URL doesn't exist :/", Status: NewFailedStatus()})
		return
	}

	var u User
	if err := models.RestoreByID(&u, bson.ObjectIdHex(t.Claims["User"].(string))); err != nil {
		e.Encode(Response{Message: "A user with that id doesnt exist!", Status: NewFailedStatus()})
		return
	}

	rev := Review{ID: bson.NewObjectId(), Content: review, From: u.ID, Repository: rep.ID}
	if err := models.Persist(rev); err != nil {
		e.Encode(Response{Message: "Could not insert that review!", Status: NewServerErrorStatus()})
		return
	}

	e.Encode(Response{Message: "Congrats you made a review!", Status: NewOKStatus(), Data: rev})
}

// GetReviewHandler retrieves a Review.
// 		GET /repo/{host}/{user}/{name}/{review}
func GetReviewHandler(w http.ResponseWriter, r *http.Request) {
	//TODO
}
