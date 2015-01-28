// repo-review is an open source project for rating and comparing hosted git repositories.
// Written by Harrison Shoebridge (http://harrisonshoebridge.me) available under the MIT license.
// Please contribute :) It makes me happy!
package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
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
	session            *mgo.Session
	verifyKey, signKey []byte
	signingMethod      jwt.SigningMethod

	usernameAndPasswordRegex *regexp.Regexp // Compiled regex for quicker matching.
	emailRegex               *regexp.Regexp // Compiled regex for quicker matching.
)

// User is someone who has registered on the site.
type User struct {
	ID           bson.ObjectId `bson:"_id" json:"_id"`
	Username     string        `bson:"username" json:"username"`
	PasswordHash string        `bson:"password_hash" json:"-"`
	Email        string        `bson:"email" json:"email"`
	PasswordSalt string        `bson:"password_salt" json:"-"`
}

// A Review is created by a User to express their feelings of a particular Repository
type Review struct {
	ID         bson.ObjectId `bson:"_id" json:"_id"`
	From       bson.ObjectId `bson:"from" json:"from"`
	Repository bson.ObjectId `bson:"repository" json:"repository"`
	Content    string        `bson:"content" json:"content"`
	Rating     int           `bson:"rating" json:"rating"`
}

// Repository is the representation of a git project on Rr
type Repository struct {
	ID   bson.ObjectId `bson:"_id" json:"_id"`
	Host string        `bson:"host" json:"host"`
	User string        `bson:"user" json:"user"`
	Name string        `bson:"name" json:"name"`
}

// Token is a container used to send a User their access_token
type Token struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Response is used as a general response for JSON rest requests
type Response struct {
	Status  Status      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Status represents a general http error
type Status struct {
	Code    int    `bson:"code" json:"code"`
	Message string `bson:"message" json:"message"`
	Error   bool   `bson:"error" json:"error"`
}

// NewOKStatus returns a new Status object with no errors.
func NewOKStatus() Status {
	return Status{http.StatusOK, "Everything is awesome!", false}
}

// NewFailedStatus returns a new Status object saying a request failed.
func NewFailedStatus() Status {
	return Status{http.StatusConflict, "Well this is awkward...", true}
}

// NewForbiddenStatus returns a new Status object detailing a failure of authorization.
func NewForbiddenStatus() Status {
	return Status{http.StatusForbidden, "You can't go here :)", true}
}

// NewServerErrorStatus returns a new Status object saying a server error has occured
func NewServerErrorStatus() Status {
	return Status{http.StatusInternalServerError, "Something bad has happened, we're sending the calvalry.", true}
}

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
	var err error

	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	defer session.Close()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/user/create", Headers(NewUserHandler)).Methods("POST")

	api.HandleFunc("/user/login", Headers(LoginUserHandler)).Methods("POST")

	api.HandleFunc("/user", Headers(Restrict(GetCurrentUserHandler))).Methods("GET")

	api.HandleFunc("/user/{username}", Headers(GetUserHandler)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}/review", Headers(Restrict(NewReviewHandler))).Methods("POST")

	api.HandleFunc("/repo/{host}/{user}/{name}/{review}", Headers(GetReviewHandler)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}", Headers(GetRepository)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}", Headers(Restrict(NewRepository))).Methods("POST")

	r.HandleFunc("/secret", Headers(Restrict(GetSecretHandler))).Methods("GET")

	// Serve ALL the static files!
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", r)

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
	username, email, password := r.FormValue("username"), r.FormValue("email"), r.FormValue("password")
	uRe, eRe, pRe := usernameAndPasswordRegex.FindString(username), emailRegex.FindString(email), usernameAndPasswordRegex.FindString(username)
	e := json.NewEncoder(w)

	if uRe == "" || eRe == "" || pRe == "" {
		e.Encode(Response{Message: "Your username, password or email is not valid.", Status: NewFailedStatus()})
		return
	}

	c := session.DB(db).C("users")
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

	c := session.DB(db).C("users")
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

	c := session.DB(db).C("users")

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

	c := session.DB(db).C("users")
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

	c := session.DB(db).C("repositories")
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

	c := session.DB(db).C("repositories")
	var rep Repository

	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&rep); rep == (Repository{}) {
		e.Encode(Response{Message: "A repo with that URL doesn't exist :/", Status: NewFailedStatus()})
		return
	}

	c = session.DB(db).C("users")
	var u User

	if c.Find(bson.M{"_id": bson.ObjectIdHex(t.Claims["User"].(string))}).One(&u); u == (User{}) {
		e.Encode(Response{Message: "A user with that id doesnt exist!", Status: NewFailedStatus()})
		return
	}

	c = session.DB(db).C("reviews")
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

}

// GetRepository retrieves a Repository.
// 		GET /api/repo/{host}/{user}/{name}
func GetRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]
	e := json.NewEncoder(w)

	c := session.DB(db).C("repositories")
	var re Repository

	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re == (Repository{}) {
		e.Encode(Response{Message: "That repository does not exist", Status: NewFailedStatus()})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Here is your repo", Status: NewOKStatus(), Data: re})
}

// Headers adds JSON headers onto a request.
func Headers(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fn(w, r)
	}
}

// Restrict checks if a provided access_token is valid, if it is continue the request.
func Restrict(fn func(http.ResponseWriter, *http.Request, *jwt.Token)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.FormValue("access_token")
		e := json.NewEncoder(w)

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

		if err != nil {
			e.Encode(Response{Message: "That is not a valid token", Status: NewFailedStatus()})
			fmt.Println(err)
			return
		}

		if !token.Valid {
			e.Encode(Response{Message: "Something obsurely strange happened to your token", Status: NewServerErrorStatus()})
		}

		fn(w, r, token)
	}
}
