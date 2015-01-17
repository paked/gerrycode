package main

import (
	"fmt"
	// "github.com/freehaha/token-auth"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type User struct {
	Id           bson.ObjectId `bson:"_id"`
	Username     string        `bson:"username",json:"username"`
	PasswordHash string        `bson:"password_hash",json:"password_hash"`
	PasswordSalt string
	Email        string
}

type Review struct {
	From bson.ObjectId `bson:"_id"`
}

var (
	session *mgo.Session

	db = "repo-reviews"
)

func main() {
	var err error

	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	// defer session.Close()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// POST /api/user/create?username=paked&pasword=pw
	// Create new user
	api.HandleFunc("/user/create", newUserHandler).Methods("POST")

	// POST /api/user/login?username=paked&password=pw
	// Authenticate and return token
	api.HandleFunc("/user/login", loginUserHandler).Methods("POST")

	// GET /api/user?api_token=xxx
	// Return the current user
	api.HandleFunc("/user", getCurrentUserHandler).Methods("GET")

	// GET /api/user/{username}
	// Return the specified user (if they exist)
	api.HandleFunc("/user/{username}", getUserHandler).Methods("GET")

	// POST /api/repo/{repository}/review?text=This+sucks&rating=2&api_token=xxx
	// Submit a new review
	api.HandleFunc("/repo/{repository}/review", newReviewHandler).Methods("POST")

	// GET /repo/{repository}/{review}
	// Return a review from a repository
	api.HandleFunc("/repo/{repository}/{review}", getReviewHandler).Methods("GET")

	// GET /api/repo/{repository}
	// Get information and all the reviews on a repo
	api.HandleFunc("/repo/{repository}", getRepository).Methods("GET")
	http.Handle("/", r)

	fmt.Println("Loading http server on :8080...")

	http.ListenAndServe(":8080", nil)

}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	username, password := r.FormValue("username"), r.FormValue("password")

	if username == "" || password == "" {
		fmt.Fprintln(w, "Username or password is not valid")
		return
	}

	c := session.DB(db).C("users")
	var u User

	if err := c.Find(bson.M{"username": username}).One(&u); err != nil {
		fmt.Println(err)
		return
	}

	if u != (User{}) {
		fmt.Fprint(w, "That user already exists!")
		return
	}

	u = User{Id: bson.NewObjectId(), Username: username, PasswordHash: password}

	if err := c.Insert(u); err != nil {
		panic(err)
		return
	}

	fmt.Fprintf(w, "%v", u)

}

func getUserHandler(w http.ResponseWriter, r *http.Request) {

}

func loginUserHandler(w http.ResponseWriter, r *http.Request) {
	username, password := r.FormValue("username"), r.FormValue("password")

	if username == "" || password == "" {
		fmt.Fprintln(w, "your username and password don't have anything in them")
		return
	}

	c := session.DB(db).C("users")

	var u User

	if err := c.Find(bson.M{"username": username, "password_hash": password}).One(&u); err != nil {
		panic(err)
		return
	}

	if u == (User{}) {
		fmt.Fprintln(w, "That user doesnt exist")
	}

	fmt.Fprintf(w, "yes you are %v", u.Username)

}

func getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {

}

func newReviewHandler(w http.ResponseWriter, r *http.Request) {

}

func getReviewHandler(w http.ResponseWriter, r *http.Request) {

}

func getRepository(w http.ResponseWriter, r *http.Request) {

}
