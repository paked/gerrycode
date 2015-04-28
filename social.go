package main

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/paked/models"
	"github.com/sqs/mux"
	"gopkg.in/mgo.v2/bson"
)

type Follow struct {
	ID   bson.ObjectId `bson: "_id" json:"id"`
	To   bson.ObjectId `bson:"to" json:"to"`
	From bson.ObjectId `bson:"from" json:"from"`
}

func (f Follow) BID() bson.ObjectId {
	return f.ID
}

func (f Follow) C() string {
	return "follows"
}

func PostFollowUserHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	c := NewCommunicator(w)
	vars := mux.Vars(r)
	user := vars["id"]
	toFollow := vars["follow"]
	var f Follow

	if !bson.IsObjectIdHex(user) || !bson.IsObjectIdHex(toFollow) {
		c.Fail("Those are not valid ObjectIds!")
		return
	}

	if err := models.Restore(&f, bson.M{"to": bson.ObjectIdHex(user)}); err != nil {
		c.Error("You are already following that user!")
		return
	}

	f = Follow{
		ID:   bson.NewObjectId(),
		To:   bson.ObjectIdHex(toFollow),
		From: bson.ObjectIdHex(user),
	}

	if err := models.Persist(f); err != nil {
		c.Fail("Somethign went wrong in the database")
		return
	}

	c.OKWithData("Everything is A OK.", f)
}
