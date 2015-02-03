package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"regexp"
	"time"
)

const (
	usernameAndPasswordRegexString = `^[a-zA-Z]\w*[a-zA-Z]$` // 1st and last characters must be letters.
	emailRegexString               = `^.*\@.*$`              // As long as it has an '@' symbol in it I don't care.
)

var (
	usernameAndPasswordRegex *regexp.Regexp
	emailRegex               *regexp.Regexp
)

// User is someone who has registered on the site.
type User struct {
	ID           bson.ObjectId `bson:"_id" json:"_id"`
	Username     string        `bson:"username" json:"username"`
	PasswordHash string        `bson:"password_hash" json:"-"`
	Email        string        `bson:"email" json:"email"`
	PasswordSalt string        `bson:"password_salt" json:"-"`
}

func (u User) BID() bson.ObjectId {
	return u.ID
}

func (u User) C() string {
	return "users"
}

func LoginUser(username string, password string) (bool, User, error) {
	u := User{}

	c := server.Collection(u.C())
	if err := c.Find(bson.M{"username": username, "password": password}).One(&u); err != nil && u == (User{}) {
		return false, User{}, err
	}

	return true, u, nil
}

// NewUserHandler creates a new user.
// 		POST /api/user/create?username=paked&pasword=pw
func NewUserHandler(w http.ResponseWriter, r *http.Request) {
	username := usernameAndPasswordRegex.FindString(r.FormValue("username"))
	email := emailRegex.FindString(r.FormValue("email"))
	password := usernameAndPasswordRegex.FindString("password")

	e := json.NewEncoder(w)

	if username == "" || email == "" || password == "" {
		e.Encode(Response{Message: "Your username, password or email is not valid.", Status: NewFailedStatus()})
		return
	}

	c := server.Collection("users")

	var u User
	if c.Find(bson.M{"username": username}).One(&u); u != (User{}) {
		e.Encode(Response{Message: "That user already exists!", Status: NewFailedStatus()})
		return
	}

	u = User{ID: bson.NewObjectId(), Username: username, Email: email, PasswordHash: password}
	if err := c.Insert(u); err != nil {
		e.Encode(Response{Message: "Could not submit that user", Status: NewServerErrorStatus()})
		return
	}

	e.Encode(Response{Message: "Here is your user!", Status: NewOKStatus(), Data: u})
}

// GetUserHandler retrieves a User from the database
// 		GET /api/user/{username}
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e := json.NewEncoder(w)

	c := server.Collection("users")

	var u User
	if c.Find(bson.M{"username": vars["username"]}).One(&u); u == (User{}) {
		e.Encode(Response{Message: "That user does not exist", Status: NewFailedStatus()})
		return
	}

	e.Encode(Response{Message: "We found that user!", Status: NewOKStatus(), Data: u})
}

// LoginUserHandler checks the provided login credentials and if valid return an access_token.
//		POST /api/user/login?username=paked&password=pw
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	username, password := r.FormValue("username"), r.FormValue("password")
	e := json.NewEncoder(w)

	if username == "" || password == "" {
		e.Encode(Response{Message: "That is not a valid username or password", Status: NewFailedStatus()})
		return
	}

	res, u, err := LoginUser(username, password)

	if err != nil || !res {
		e.Encode(Response{Message: "Could not find your user :)", Status: NewFailedStatus()})
		return
	}

	t := jwt.New(signingMethod)

	t.Claims["AccessToken"] = "1"
	t.Claims["User"] = u.ID
	t.Claims["Expires"] = time.Now().Add(time.Minute * 15).Unix()

	tokenString, err := t.SignedString(signKey)

	if err != nil {
		e.Encode(Response{Message: "We could not sign the token made for you", Status: NewServerErrorStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is your token!", Status: NewOKStatus(), Data: tokenString})
}

// GetCurrentUserHandler retrieves the User currently logged in.
// 		GET /api/user?api_token=xxx
func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	id, ok := t.Claims["User"].(string)
	e := json.NewEncoder(w)

	if !ok {
		e.Encode(Response{Message: "Could not cast interface to that bson.ObjectId!", Status: NewServerErrorStatus()})
		return
	}

	c := server.Collection("users")

	var u User
	if c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&u); u == (User{}) {
		e.Encode(Response{Message: "Could not find that user!", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is you!", Status: NewOKStatus(), Data: u})
}
