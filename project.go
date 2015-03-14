package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
)

type Project struct {
	ID    bson.ObjectId `bson:"_id"`
	Owner bson.ObjectId `bson:"owner"`
	Name  string        `bson:"name"`
	URL   string        `bson:"url"`
	TLDR  string        `bson:"tldr"`
}

func (p Project) BID() bson.ObjectId {
	return p.ID
}

func (p Project) C() string {
	return "projects"
}

type Flag struct {
	ID      bson.ObjectId `bson:"_id"`
	Project bson.ObjectId `bson:"project"`
	Query   string        `bson:"query"`
	Time    time.Time     `bson:"time"`
}

func (f Flag) BID() bson.ObjectId {
	return f.BID()
}

func (f Flag) C() string {
	return "flags"
}

type Feedback struct {
	ID      bson.ObjectId `bson:"_id"`
	Project bson.ObjectId `bson:"project"`
	Flag    bson.ObjectId `bson:"flag"`
	Text    bson.ObjectId `bson:"text"`
}

func (f Feedback) BID() bson.ObjectId {
	return f.BID()
}

func (f Feedback) C() string {
	return "feedback"
}

func PostCreateProject(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)
	var p Project
	name, url, tldr := r.FormValue("name"), r.FormValue("url"), r.FormValue("tldr")
	id, ok := t.Claims["User"].(string)

	if !ok {
		e.Encode(Response{Message: "Unable to get that user... logout maybe?", Status: NewFailedStatus()})
		return
	}

	if err := models.Restore(&p, bson.M{"url": url}); err == nil {
		e.Encode(Response{Message: "That project already exists", Status: NewFailedStatus()})
		return
	}

	p = Project{ID: bson.NewObjectId(), Owner: bson.ObjectIdHex(id), Name: name, URL: url, TLDR: tldr}
	if err := models.Persist(p); err != nil {
		e.Encode(Response{Message: "Error persisting your new project", Status: NewFailedStatus()})
		return
	}

	e.Encode(Response{Message: "Here is the project", Status: NewOKStatus()})
}
