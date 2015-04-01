package main

import (
	"net/http"

	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
)

func GetTopProjects(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)

	var projects []Project
	project := Project{}
	iter, err := models.Fetch(project.C(), bson.M{}, "-time")
	if err != nil {
		c.Fail("Unable to fetch projects...")
	}

	var counter int
	for iter.Next(&project) && counter < 10 {
		projects = append(projects, project)
		counter += 1
	}

	c.OKWithData("Here are the top projects", projects)
}
