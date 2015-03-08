package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/github"

	"net/http"
)

func GetUsersRepositories(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)
	client := github.NewClient(nil)

	repos, _, err := client.Repositories.List("paked", nil)

	if err != nil {
		fmt.Println(err)
	}

	e.Encode(repos)
}
