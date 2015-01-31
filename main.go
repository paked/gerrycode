// repo-review is an open source project for rating and comparing hosted git repositories.
// Written by Harrison Shoebridge (http://github.com/paked) available under the MIT license.
//
// Please contribute :) It makes me happy!
package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

const (
	usernameAndPasswordRegexString = `^[a-zA-Z]\w*[a-zA-Z]$` // 1st and last characters must be letters.
	emailRegexString               = `^.*\@.*$`              // As long as it has an '@' symbol in it I don't care.

	db             = "repo-reviews" // Mongodb database name
	privateKeyPath = "app.rsa"      // Command: openssl genrsa -out app.rsa 1024
	publicKeyPath  = "app.rsa.pub"  // Command: openssl rsa -in app.rsa -pubout > app.rsa.pub
)

var (
	server             *Server
	verifyKey, signKey []byte
	signingMethod      jwt.SigningMethod

	usernameAndPasswordRegex *regexp.Regexp // Compiled regex for quicker matching.
	emailRegex               *regexp.Regexp // Compiled regex for quicker matching.
)

func init() {
	var err error

	signKey, err = ioutil.ReadFile(privateKeyPath)

	if err != nil {
		fmt.Println("Could not find your private key!")
		panic(err)
	}

	verifyKey, err = ioutil.ReadFile(publicKeyPath)

	if err != nil {
		fmt.Println("Could not find your public key!")
		panic(err)
	}

	signingMethod = jwt.GetSigningMethod("RS256")

	usernameAndPasswordRegex, err = regexp.Compile(usernameAndPasswordRegexString)

	if err != nil {
		panic(err)
	}

	emailRegex, err = regexp.Compile(emailRegexString)

	if err != nil {
		panic(err)
	}

}

func main() {
	server = NewServer()

	fmt.Println("Loading http server on :8080...")

	fmt.Println(http.ListenAndServe(":8080", nil))

}

// GetSecretHandler is a test handler to check if access_tokens work.
// 		GET /secret
func GetSecretHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	json.NewEncoder(w).Encode(Response{Message: "NCSS IS ILLUMINATTI", Status: NewOKStatus()})
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

	c := server.Conn.DB(db).C("users")

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

	c := server.Conn.DB(db).C("users")

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

	c := server.Conn.DB(db).C("users")

	var u User
	if c.Find(bson.M{"username": username, "password_hash": password}).One(&u); u == (User{}) {
		e.Encode(Response{Message: "A user with that username and password combination does not exist.", Status: NewFailedStatus()})
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

	c := server.Conn.DB(db).C("users")

	var u User
	if c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&u); u == (User{}) {
		e.Encode(Response{Message: "Could not find that user!", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is you!", Status: NewOKStatus(), Data: u})
}

// NewRepository creates a new Repository link.
// 		POST /api/repo/{host}/{user}/{name}?access_token=xxx
func NewRepository(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]
	e := json.NewEncoder(w)

	c := server.Conn.DB(db).C("repositories")

	var re Repository
	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re != (Repository{}) {
		e.Encode(Response{Message: "That repo already exists", Status: NewFailedStatus()})
		return
	}

	re = Repository{ID: bson.NewObjectId(), Host: host, User: user, Name: name}
	if err := c.Insert(re); err != nil {
		e.Encode(Response{Message: "Could not insert that repository", Status: NewServerErrorStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "We created that new repository!", Status: NewOKStatus(), Data: re})
}

// NewReviewHandler creates a new Review on a Repository.
// 		POST /api/repo/{repository}/review?text=This+sucks&rating=2&access_token=xxx
func NewReviewHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name, review := vars["host"], vars["user"], vars["name"], r.FormValue("review")
	e := json.NewEncoder(w)

	if review == "" {
		e.Encode(Response{Message: "Please say something in your review :)", Status: NewFailedStatus()})
		return
	}

	c := server.Conn.DB(db).C("repositories")

	var rep Repository
	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&rep); rep == (Repository{}) {
		e.Encode(Response{Message: "A repo with that URL doesn't exist :/", Status: NewFailedStatus()})
		return
	}

	c = server.Conn.DB(db).C("users")

	var u User
	if c.Find(bson.M{"_id": bson.ObjectIdHex(t.Claims["User"].(string))}).One(&u); u == (User{}) {
		e.Encode(Response{Message: "A user with that id doesnt exist!", Status: NewFailedStatus()})
		return
	}

	c = server.Conn.DB(db).C("reviews")

	rev := Review{ID: bson.NewObjectId(), Content: review, From: u.ID, Repository: rep.ID}
	if err := c.Insert(rev); err != nil {
		e.Encode(Response{Message: "Could not insert that review!", Status: NewServerErrorStatus()})
		return
	}

	e.Encode(Response{Message: "Congrats you made a review!", Status: NewOKStatus(), Data: rev})
}

// GetReviewHandler retrieves a Review.
// 		GET /repo/{host}/{user}/{name}/{review}
func GetReviewHandler(w http.ResponseWriter, r *http.Request) {
	//TODO
}

// GetRepository retrieves a Repository.
// 		GET /api/repo/{host}/{user}/{name}
func GetRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]
	e := json.NewEncoder(w)

	c := server.Conn.DB(db).C("repositories")

	var re Repository
	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re == (Repository{}) {
		e.Encode(Response{Message: "That repository does not exist", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is your repo", Status: NewOKStatus(), Data: re})
}
