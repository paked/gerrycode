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

	transport, err := transport(bson.ObjectIdHex(t.Claims["User"].(string)))
	if err != nil {
		e.Encode(Response{Message: "Error creating transport", Status: NewFailedStatus()})
		return
	}

	client := github.NewClient(transport.Client())

	repos, _, err := client.Repositories.List("paked", nil)

	if err != nil {
		fmt.Println(err)
	}

	e.Encode(repos)
}

func PostLinkUserAccount(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	e := json.NewEncoder(w)

	id, ok := t.Claims["User"].(string)
	if !ok {
		e.Encode(Response{Message: "Unable to get that id :/", Status: NewFailedStatus()})
		return
	}

	fmt.Println(conf.ClientID)

	session, err := store.Get(r, "github-auth")
	if err != nil {
		e.Encode(Response{Message: "Unable to get that session :/", Status: NewFailedStatus()})
		return
	}
	session.AddFlash(id)
	if err := session.Save(r, w); err != nil {
		e.Encode(Response{Message: "Unable to save that session!", Status: NewFailedStatus()})
	}

	http.Redirect(w, r, oauthConfig.AuthCodeURL(""), http.StatusFound)
}

func GetAuthedGithubAccount(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	session, err := store.Get(r, "github-auth")
	if err != nil {
		e.Encode(Response{Message: "Unable to get that session :/", Status: NewFailedStatus()})
		return
	}

	flashes := session.Flashes()
	if len(flashes) == 0 {
		e.Encode(Response{Message: "You got something funky going on with your cookies!", Status: NewFailedStatus()})
		return
	}

	id, ok := flashes[0].(string)
	if !ok {
		e.Encode(Response{Message: "Unable to get that id :/", Status: NewFailedStatus()})
		return
	}

	t, err := firstTransport(r.FormValue("code"))
	if err != nil {
		e.Encode(Response{Message: "Error creating transport", Status: NewFailedStatus()})
		return
	}

	la := LinkedAccount{ID: bson.NewObjectId(),
		Origin:  bson.ObjectIdHex(id),
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

	if err := models.Restore(&la, bson.M{"origin": id}); err != nil {
		return nil, err
	}

	return &oauth.Transport{Token: &oauth.Token{AccessToken: la.Token}}, nil
}
