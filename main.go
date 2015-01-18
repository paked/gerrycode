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
	"time"
)

type User struct {
	Id           bson.ObjectId `bson:"_id"`
	Username     string        `bson:"username",json:"username"`
	PasswordHash string        `bson:"password_hash",json:"password_hash"`
	PasswordSalt string
	Email        string
}

type Review struct {
	Id      bson.ObjectId `bson:"_id"`
	From    bson.ObjectId `bson:"from"`
	Content string        `bson:"content"`
	Rating  int           `bson:"rating"`
}

type Token struct {
	Name  string
	Value string
}

var (
	session            *mgo.Session
	verifyKey, signKey []byte
	signingMethod      jwt.SigningMethod
)

const (
	db             = "repo-reviews"
	privateKeyPath = "app.rsa"
	publicKeyPath  = "app.rsa.pub"
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
}

func NewAccessToken(value string) Token {
	return Token{"AccessToken", value}
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

	// GET /secret
	// A page to test secrecy!
	r.HandleFunc("/secret", restrict(getSecret)).Methods("GET")
	http.Handle("/", r)

	fmt.Println("Loading http server on :8080...")

	http.ListenAndServe(":8080", nil)

}

func getSecret(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "NCSS IS ILLUMINATTI")
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {

}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	c := session.DB(db).C("users")
	var u User

	if c.Find(bson.M{"username": vars["username"]}).One(&u); u == (User{}) {
		fmt.Fprintln(w, "that user doesnt exist")
		return
	}

	fmt.Fprintln(w, "We found that user!", u)
}

func loginUserHandler(w http.ResponseWriter, r *http.Request) {
	username, password := r.FormValue("username"), r.FormValue("password")

	if username == "" || password == "" {
		fmt.Fprintln(w, "your username and password don't have anything in them")
		return
	}

	c := session.DB(db).C("users")

	var u User

	if c.Find(bson.M{"username": username, "password_hash": password}).One(&u); u == (User{}) {
		fmt.Fprintln(w, "That user doesnt exist")
		return
	}

	t := jwt.New(signingMethod)

	t.Claims["AccessToken"] = "1"
	t.Claims["User"] = u.Id
	t.Claims["Expires"] = time.Now().Add(time.Minute * 15).Unix()

	tokenString, err := t.SignedString(signKey)

	if err != nil {
		fmt.Fprintln(w, "Error signing that token")
		return
	}

	json.NewEncoder(w).Encode(NewAccessToken(tokenString))

}

func getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {

}

func newReviewHandler(w http.ResponseWriter, r *http.Request) {

}

func getReviewHandler(w http.ResponseWriter, r *http.Request) {

}

func getRepository(w http.ResponseWriter, r *http.Request) {

}

func restrict(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.FormValue("access_token")
		fmt.Println(tokenString)

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

		if err != nil {
			fmt.Fprintln(w, "That is not a valid token")
			fmt.Println(err)
			return
		}

		if !token.Valid {
			fmt.Fprintln(w, "Something obscurely strange happened to uyour token")
		}

		fmt.Println("WE GAVE A TOKEN ACCESS TO SOMETHING!")
		fn(w, r)

	}
}
