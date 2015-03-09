package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/github"
	"gopkg.in/mgo.v2/bson"
)

const (
	GithubAccount = "github"
)

func fillOAuthConfig() {
	oauthConfig = &oauth.Config{
		ClientId:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Scope:        "(no scope)",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		RedirectURL:  "http://localhost:8080/api/oauth",
	}

}

type LinkedAccount struct {
	ID       bson.ObjectId `bson:"id"`
	Origin   bson.ObjectId `bson:"origin"`
	Service  string        `bson:"service"`
	Username string        `bson:"username"`
}

func (la LinkedAccount) BID() bson.ObjectId {
	return la.ID
}

func (la LinkedAccount) C() string {
	return "links"
}

func GetUsersRepositories(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)
	client := github.NewClient(nil)

	repos, _, err := client.Repositories.List("paked", nil)

	if err != nil {
		fmt.Println(err)
	}

	e.Encode(repos)
}

func PostLinkUserAccount(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	//	id, ok := mux.Vars(r)["user"]
	e := json.NewEncoder(w)

	if false {
		e.Encode(Response{Message: "Unable to get that id :/", Status: NewFailedStatus()})
		return
	}
	fmt.Println(conf.ClientID)

	http.Redirect(w, r, oauthConfig.AuthCodeURL("gooey"), http.StatusFound)
}

func GetAuthedGithubAccount(w http.ResponseWriter, r *http.Request) {
	t := &oauth.Transport{Config: oauthConfig}
	t.Exchange(r.FormValue("code"))
	// The Transport now has a valid Token. Create an *http.Client
	// with which we can make authenticated API requests.
	c := t.Client()
	fmt.Println(c)
	fmt.Fprintf(w, "woo logged in!")
	//c.Post(...)
	// ...
	// btw, r.FormValue("state") == "foo"
}
