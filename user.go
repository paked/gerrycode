package main

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"github.com/paked/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

const (
	credentialsRaw = `^[a-zA-Z]\w*[a-zA-Z]$` // 1st and last characters must be letters.
	emailRaw       = `^.*\@.*$`              // As long as it has an '@' symbol in it I don't care.
)

var (
	credentialsRegex *regexp.Regexp
	emailRegex       *regexp.Regexp
)

// User is someone who has registered on the site.
type User struct {
	ID           bson.ObjectId `bson:"_id" json:"_id"`
	Username     string        `bson:"username" json:"username"`
	PasswordHash []byte        `bson:"password_hash" json:"-"`
	Email        string        `bson:"email" json:"email"`
}

func (u User) BID() bson.ObjectId {
	return u.ID
}

func (u User) C() string {
	return "users"
}

func LoginUser(username string, password string) (User, error) {
	u := User{}
	if err := models.Restore(&u, bson.M{"username": username}); err != nil {
		return u, err
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return u, errors.New("Those passwords don't match")
	}

	return u, nil
}

func ValidateCredentials(username, password string) bool {
	if !credentialsRegex.MatchString(username) {
		return false
	}

	if password == "" {
		return false
	}

	return true
}

// NewUserHandler creates a new user.
// 		POST /api/user/register?username=paked&pasword=pw
func NewUserHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !ValidateCredentials(username, password) || !emailRegex.MatchString(email) {
		c.Fail("That is not a valid username password, or email")
	}
	var u User
	if err := models.Restore(&u, bson.M{"username": username}); err == nil {
		c.Fail("That user already exists!")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		c.Error("Something bad happened while hashing your password!")
		return
	}

	u = User{ID: bson.NewObjectId(), Username: username, Email: email, PasswordHash: passwordHash}
	if err := models.Persist(u); err != nil {
		c.Error("Unable to persist that new user!")
		return
	}

	c.OKWithData("Here is your new user", u)
}

// GetUserHandler retrieves a User from the database
// 		GET /api/user/{username}
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)
	vars := mux.Vars(r)
	lookup := vars["username"]

	var err error
	var u User
	if bson.IsObjectIdHex(lookup) {
		err = models.RestoreByID(&u, bson.ObjectIdHex(lookup))
	} else {
		err = models.Restore(&u, bson.M{"username": lookup})
	}

	if err != nil {
		c.Fail("Can't find that user")
		return
	}

	c.OKWithData("Here is the user you were looking for", u)
}

// LoginUserHandler checks the provided login credentials and if valid return an access_token.
//		POST /api/user/login?username=paked&password=pw
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	c := NewCommunicator(w)
	username := credentialsRegex.FindString(r.FormValue("username"))
	password := credentialsRegex.FindString(r.FormValue("password"))

	if username == "" || password == "" {
		c.Fail("That is not a valid username or password")
		return
	}

	u, err := LoginUser(username, password)

	if err != nil {
		c.Fail("Could not login your user")
		return
	}

	t := jwt.New(signingMethod)

	t.Claims["AccessToken"] = "1"
	t.Claims["User"] = u.ID
	t.Claims["exp"] = time.Now().Add(time.Hour * 12).Unix()

	tokenString, err := t.SignedString(signKey)

	if err != nil {
		c.Error("Error signing that token")
		return
	}

	c.OKWithData("Here is your token", tokenString)
}

// GetCurrentUserHandler retrieves the User currently logged in.
// 		GET /api/user?api_token=xxx
func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	c := NewCommunicator(w)
	id, ok := t.Claims["User"].(string)

	if !ok {
		c.Error("Could not cast that id to string!")
		return
	}

	var u User
	if err := models.RestoreByID(&u, bson.ObjectIdHex(id)); err != nil {
		c.Fail("Could not find that user!")
		return
	}

	c.OKWithData("Here is you", u)
}
