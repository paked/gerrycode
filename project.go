package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
)

// Project represents a project which a User has submitted
type Project struct {
	ID    bson.ObjectId `bson:"_id" json:"id"`
	Owner bson.ObjectId `bson:"owner" json:"owner"`
	Name  string        `bson:"name" json:"name"`
	URL   string        `bson:"url" json:"url"`
	TLDR  string        `bson:"tldr" json:"tldr"`
}

// BID is a helper function to fulfill the models.Modeller interface
func (p Project) BID() bson.ObjectId {
	return p.ID
}

// C is a helper function to fulfill the models.Modeller interface
func (p Project) C() string {
	return "projects"
}

// Flag represents a flag by the project owner requesting feedback
type Flag struct {
	ID      bson.ObjectId `bson:"_id" json:"id"`
	Project bson.ObjectId `bson:"project" json:"project"`
	Query   string        `bson:"query" json:"query"`
	Time    time.Time     `bson:"time" json:"time"`
}

// BID a helper function to fulfill the models.Modeller interface
func (f Flag) BID() bson.ObjectId {
	return f.BID()
}

// C a helper function to fulfill the models.Modeller interface
func (f Flag) C() string {
	return "flags"
}

// Feedback represents feedback given by a User on a "flagged" change
type Feedback struct {
	ID      bson.ObjectId `bson:"_id" json:"id"`
	Project bson.ObjectId `bson:"project" json:"project"`
	Flag    bson.ObjectId `bson:"flag" json:"flag"`
	Text    string        `bson:"text" json:"text"`
	User    bson.ObjectId `bson:"user" json:"user"`
}

// BID a helper function to fulfill the models.Modeller interface
func (f Feedback) BID() bson.ObjectId {
	return f.BID()
}

// C a helper function to fulfill the models.Modeller interface
func (f Feedback) C() string {
	return "feedback"
}

// PostCreateProject is the handler to create a project
func PostCreateProjectHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
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

	e.Encode(Response{Message: "Here is the project", Status: NewOKStatus(), Data: p})
}

// GetRepository retrieves a Repository.
// 		GET /api/project/{id}
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e := json.NewEncoder(w)
	id := vars["id"]

	var project Project
	if err := models.RestoreByID(&project, bson.ObjectIdHex(id)); err != nil {
		e.Encode(Response{Message: "That project does not exist", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is your project", Status: NewOKStatus(), Data: project})
}

func PostFlagForFeedbackHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)
	query := r.FormValue("query")
	project := mux.Vars(r)["id"]

	f := Flag{ID: bson.NewObjectId(), Query: query, Project: bson.ObjectIdHex(project), Time: time.Now()}
	if err := models.Persist(f); err != nil {
		e.Encode(Response{Message: "Could not persist project!", Status: NewFailedStatus()})
		return
	}

	e.Encode(Response{Message: "Here is your new flag...", Status: NewOKStatus(), Data: f})
}

// GetUsersProjectsHandler gets the current users projects and returns them in a JSON object
func GetUsersProjectsHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)

	id, ok := t.Claims["User"].(string)
	if !ok {
		e.Encode(Response{Message: "Could not cast interface to a string!", Status: NewServerErrorStatus()})
		return
	}

	var projects []Project
	project := Project{}
	iter, err := models.Fetch(project.C(), bson.M{"owner": bson.ObjectIdHex(id)})
	if err != nil {
		e.Encode(Response{Message: "Something went wrong fetching projects...", Status: NewServerErrorStatus()})
		return
	}

	for iter.Next(&project) {
		projects = append(projects, project)
	}

	e.Encode(Response{Message: "Here are your projects!", Status: NewOKStatus(), Data: projects})
}

func GetProjectsFlagsHandler(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)

	var flags []Flag
	flag := Flag{}
	iter, err := models.Fetch(flag.C(), bson.M{"project": bson.ObjectIdHex(mux.Vars(r)["id"])})
	if err != nil {
		e.Encode(Response{Message: "Something went wrong fetching flags...", Status: NewServerErrorStatus()})
		return
	}

	for iter.Next(&flag) {
		flags = append(flags, flag)
	}

	e.Encode(Response{Message: "Here are your flags!", Status: NewOKStatus(), Data: flags})
}

func PostFeedbackOnFlag(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)
	vars := mux.Vars(r)
	flag := vars["flag"]
	project := vars["id"]
	userString, ok := t.Claims["User"].(string)
	if !ok {
		e.Encode(Response{Message: "Unable to cast to string...", Status: NewFailedStatus()})
		return
	}
	user := bson.ObjectIdHex(userString)

	var f Flag
	if err := models.Restore(&f, bson.M{"_id": bson.ObjectIdHex(flag), "project": bson.ObjectIdHex(project)}); err != nil {
		e.Encode(Response{Message: "unable to find that flag", Status: NewFailedStatus()})
		return
	}

	fee := Feedback{ID: bson.NewObjectId(), Flag: f.ID, Project: f.Project, Text: r.FormValue("text"), User: user}
	if err := models.Persist(fee); err != nil {
		fmt.Println(err, "feedback:", fee, "flag:", f)
		e.Encode(Response{Message: "Couldnt persist that feedback :/", Status: NewFailedStatus()})
		return
	}

	e.Encode(Response{Message: "Here is your happy feedback!", Status: NewOKStatus(), Data: fee})
}

func GetAllFeedbackForFlag(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	vars := mux.Vars(r)
	flag := bson.ObjectIdHex(vars["flag"])
	project := bson.ObjectIdHex(vars["id"])

	var feedbacks []Feedback
	feedback := Feedback{}
	iter, err := models.Fetch(feedback.C(), bson.M{"flag": flag, "project": project})
	if err != nil {
		e.Encode(Response{Message: "Something went wrong fetching flags...", Status: NewServerErrorStatus()})
		return
	}

	for iter.Next(&feedback) {
		feedbacks = append(feedbacks, feedback)
	}

	e.Encode(Response{Message: "Here are your flags!", Status: NewOKStatus(), Data: feedbacks})
}
