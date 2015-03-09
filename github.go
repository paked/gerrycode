package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/github"
	"github.com/paked/models"
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
		TokenCache:   oauth.CacheFile("auth.cache"),
	}

}

type LinkedAccount struct {
	ID      bson.ObjectId `bson:"id"`
	Origin  bson.ObjectId `bson:"origin"`
	Service string        `bson:"service"`
	Token   string        `bson:"token"`
}

func (la LinkedAccount) BID() bson.ObjectId {
	return la.ID
}

func (la LinkedAccount) C() string {
	return "links"
}

func GetUsersRepositories(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)
	var la LinkedAccount

	if err := models.Restore(&la, bson.M{"origin": bson.ObjectIdHex(t.Claims["User"].(string))}); err != nil {
		e.Encode("Failed model!")
		return
	}

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
	e := json.NewEncoder(w)
	t, err := firstTransport(r.FormValue("code"))
	if err != nil {
		e.Encode(Response{Message: "Error creating transport", Status: NewFailedStatus()})
		return
	}

	la := LinkedAccount{ID: bson.NewObjectId(),
		Origin:  bson.NewObjectId(),
		Service: GithubAccount,
		Token:   t.Token.AccessToken}

	if err := models.Persist(la); err != nil {
		e.Encode(Response{Message: "Unable to persist model!", Status: NewFailedStatus()})
		return
	}

	c := t.Client()
	fmt.Println(c, t)
	fmt.Fprintf(w, "woo logged in!")
	//c.Post(...)
	// ...
	// btw, r.FormValue("state") == "foo"
}

func firstTransport(code string) (*oauth.Transport, error) {
	t := &oauth.Transport{Config: oauthConfig}
	_, err := t.Exchange(code)

	return t, err
}

func transport(id bson.ObjectId) (*oauth.Transport, error) {
	var la LinkedAccount

	if err := models.RestoreByID(&la, id); err != nil {
		return nil, err
	}

	return &oauth.Transport{Token: &oauth.Token{AccessToken: la.Token}}, nil
}
