package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
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

func (u User) WriteReview(c string, id bson.ObjectId) (Review, error) {
	rev := Review{ID: bson.NewObjectId(), From: u.ID, Repository: id, Content: c}
	return rev, models.Persist(&rev)
}

func LoginUser(username string, password string) (User, error) {
	u := User{}
	if err := models.Restore(&u, bson.M{"username": username, "password_hash": password}); err != nil {
		return u, err
	}

	return u, nil
}

// NewUserHandler creates a new user.
// 		POST /api/user/register?username=paked&pasword=pw
func NewUserHandler(w http.ResponseWriter, r *http.Request) {
	username := usernameAndPasswordRegex.FindString(r.FormValue("username"))
	email := emailRegex.FindString(r.FormValue("email"))
	password := usernameAndPasswordRegex.FindString(r.FormValue("password"))

	e := json.NewEncoder(w)

	if username == "" || email == "" || password == "" {
		e.Encode(Response{Message: "Your username, password or email is not valid.", Status: NewFailedStatus()})
		return
	}

	var u User
	if err := models.Restore(&u, bson.M{"username": username}); err == nil {
		e.Encode(Response{Message: "That user already exists!", Status: NewFailedStatus(), Data: u})
		return
	}

	u = User{ID: bson.NewObjectId(), Username: username, Email: email, PasswordHash: password}
	if err := models.Persist(u); err != nil {
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

	var u User
	if err := models.Restore(&u, bson.M{"username": vars["username"]}); err != nil {
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

	u, err := LoginUser(username, password)

	if err != nil {
		// fmt.Println(err, res, u, username, password)
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

	var u User
	if err := models.RestoreByID(&u, bson.ObjectIdHex(id)); err != nil {
		e.Encode(Response{Message: "Could not find that user!", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is you!", Status: NewOKStatus(), Data: u})
}
