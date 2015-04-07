package main

import (
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
	ID       bson.ObjectId `bson:"_id" json:"id"`
	Owner    bson.ObjectId `bson:"owner" json:"owner"`
	Name     string        `bson:"name" json:"name"`
	URL      string        `bson:"url" json:"url"`
	TLDR     string        `bson:"tldr" json:"tldr"`
	Language Language      `bson:"language" json:"language"`
	Time     time.Time     `bson:"time" json:"time"`
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
	Title   string        `bson:"title" json:"title"`
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
	c := NewCommunicator(w)
	var p Project
	name, url, tldr := r.FormValue("name"), r.FormValue("url"), r.FormValue("tldr")
	id, ok := t.Claims["User"].(string)

	if !ok {
		c.Error("Unable to correctly marshal that id!")
		return
	}

	if err := models.Restore(&p, bson.M{"url": url}); err == nil {
		c.Fail("That project already exists!")
		return
	}

	if !bson.IsObjectIdHex(id) {
		c.Fail("That is not a recognized bson.ObjectID")
		return
	}

	lang, err := GetOrCreateLanguage(r.FormValue("lang"))
	fmt.Println(lang, r.FormValue("lang"))
	if err != nil {
		c.Error("Unable to create that language...")
	}

	p = Project{ID: bson.NewObjectId(), Owner: bson.ObjectIdHex(id), Name: name, URL: url, TLDR: tldr, Language: lang, Time: time.Now()}
	if err := models.Persist(p); err != nil {
		c.Error("Error persisting that new project!")
		return
	}

	c.OKWithData("Here is your user!", p)
}
func PostEditProject(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	c := NewCommunicator(w)
	var p Project
	vars := mux.Vars(r)
	id := vars["id"]

	if !bson.IsObjectIdHex(id) {
		c.Fail("That is not a recognized bson.ObjectID")
		return
	}

	tldr := r.FormValue("tldr")
	name := r.FormValue("name")
	url := r.FormValue("url")

	if err := models.RestoreByID(&p, bson.ObjectIdHex(id)); err != nil {
		c.Fail("Could not find that project!")
		return
	}

	update := bson.M{}

	if tldr != "" {
		update["tldr"] = tldr
	}

	if name != "" {
		update["name"] = name
	}

	if url != "" {
		update["url"] = url
	}

	if len(update) == 0 {
		c.Fail("You didnt specify any new values!")
		return
	}

	if err := models.Update(&p, update); err != nil {
		c.Error("Error persisiting that new tldr!")
		return
	}

	c.OKWithData("Updated project!", p)
}

// GetRepository retrieves a Repository.
// 		GET /api/project/{id}
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)
	vars := mux.Vars(r)
	id := vars["id"]

	if !bson.IsObjectIdHex(id) {
		c.Fail("That is not a recognized bson.ObjectID")
		return
	}

	var p Project
	if err := models.RestoreByID(&p, bson.ObjectIdHex(id)); err != nil {
		c.Fail("That project does not exist!")
		return
	}

	c.OKWithData("Here is your new project", p)
}

func PostFlagForFeedbackHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	c := NewCommunicator(w)
	query := r.FormValue("query")
	title := r.FormValue("title")
	project := mux.Vars(r)["id"]

	if !bson.IsObjectIdHex(project) {
		c.Fail("That is not a recognized bson.ObjectID")
		return
	}

	f := Flag{ID: bson.NewObjectId(), Query: query, Title: title, Project: bson.ObjectIdHex(project), Time: time.Now()}
	if err := models.Persist(f); err != nil {
		c.Error("Unable to persist that error!")
		return
	}

	c.OKWithData("Here is your new flag", f)
}

// GetUsersProjectsHandler gets the current users projects and returns them in a JSON object
func GetUsersProjectsHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	c := NewCommunicator(w)
	id, ok := t.Claims["User"].(string)
	if !ok {
		c.Error("Unable to get that id from token...")
		return
	}

	if !bson.IsObjectIdHex(id) {
		c.Fail("That is not a recognized bson.ObjectID")
		return
	}

	var projects []Project
	project := Project{}
	iter, err := models.Fetch(project.C(), bson.M{"owner": bson.ObjectIdHex(id)}, "id")
	if err != nil {
		c.Fail("Those projects don't exist!")
		return
	}

	for iter.Next(&project) {
		projects = append(projects, project)
	}

	c.OKWithData("Here are your projects", projects)
}

// GetProjectsFlagsHandler gets all of the flags on a specific project
func GetProjectsFlagsHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)
	id := mux.Vars(r)["id"]

	if !bson.IsObjectIdHex(id) {
		c.Error("That is not a valid objectID")
		return
	}

	var flags []Flag
	flag := Flag{}
	iter, err := models.Fetch(flag.C(), bson.M{"project": bson.ObjectIdHex(id)}, "id")
	if err != nil {
		c.Fail("Those flags don't exist!")
		return
	}

	for iter.Next(&flag) {
		flags = append(flags, flag)
	}

	c.OKWithData("Here are your flags", flags)
}

// GetFlagHandler retrieves a flag from the database
func GetFlagHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)
	vars := mux.Vars(r)

	flagString := vars["flag"]
	projectString := vars["id"]

	if !(bson.IsObjectIdHex(flagString) && bson.IsObjectIdHex(projectString)) {
		c.Fail("That is not a valid objectID")
		return
	}

	flagID := bson.ObjectIdHex(flagString)
	projectID := bson.ObjectIdHex(projectString)

	var f Flag
	if err := models.Restore(&f, bson.M{"project": projectID, "_id": flagID}); err != nil {
		fmt.Println(projectID, flagID)
		c.Fail("That flag doesnt exist!")
		return
	}

	c.OKWithData("Here is your flag...", f)
}

// PostFeedbackOnFlagHandler adds feedback onto a flag
func PostFeedbackOnFlagHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	c := NewCommunicator(w)
	vars := mux.Vars(r)
	flag := vars["flag"]
	project := vars["id"]

	if !(bson.IsObjectIdHex(flag) && bson.IsObjectIdHex(project)) {
		c.Fail("That is not a valid objectID")
		return
	}

	userString, ok := t.Claims["User"].(string)
	if !ok {
		c.Error("Unable to marshal that ID!")
		return
	}
	user := bson.ObjectIdHex(userString)

	var f Flag
	if err := models.Restore(&f, bson.M{"_id": bson.ObjectIdHex(flag), "project": bson.ObjectIdHex(project)}); err != nil {
		c.Fail("Could not find that flag!")
		return
	}

	fee := Feedback{ID: bson.NewObjectId(), Flag: f.ID, Project: f.Project, Text: r.FormValue("text"), User: user}
	if err := models.Persist(fee); err != nil {
		fmt.Println(err, "feedback:", fee, "flag:", f)
		c.Error("Could not persist your flag!")
		return
	}

	c.OKWithData("Here is your feedback", fee)
}

// GetAllFeedbackForFlagHandler gets all of the feedback on a specified flag
func GetAllFeedbackForFlagHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)
	vars := mux.Vars(r)
	flagString := vars["flag"]
	projectString := vars["id"]
	if !(bson.IsObjectIdHex(flagString) && bson.IsObjectIdHex(projectString)) {
		c.Fail("That is not a valid bson.ObjectID")
		return
	}

	flag := bson.ObjectIdHex(flagString)
	project := bson.ObjectIdHex(projectString)

	var feedbacks []Feedback
	feedback := Feedback{}
	iter, err := models.Fetch(feedback.C(), bson.M{"flag": flag, "project": project}, "id")
	if err != nil {
		c.Fail("Unable to get all that feedback!")
		return
	}

	for iter.Next(&feedback) {
		feedbacks = append(feedbacks, feedback)
	}

	c.OKWithData("Here is your feedback", feedbacks)
}
