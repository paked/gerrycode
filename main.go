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
	PasswordHash string        `bson:"password_hash" json:"password_hash"`
	Email        string        `bson:"email" json:"email"`
	PasswordSalt string
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

	// POST /api/user/create?username=paked&pasword=pw
	// Create new user
	api.HandleFunc("/user/create", headers(newUserHandler)).Methods("POST")

	// POST /api/user/login?username=paked&password=pw
	// Authenticate and return token
	api.HandleFunc("/user/login", headers(loginUserHandler)).Methods("POST")

	// GET /api/user?api_token=xxx
	// Return the current user
	api.HandleFunc("/user", headers(restrict(getCurrentUserHandler))).Methods("GET")

	// GET /api/user/{username}
	// Return the specified user (if they exist)
	api.HandleFunc("/user/{username}", headers(getUserHandler)).Methods("GET")

	// POST /api/repo/{repository}/review?text=This+sucks&rating=2&access_token=xxx
	// Submit a new review
	api.HandleFunc("/repo/{host}/{user}/{name}/review", headers(restrict(newReviewHandler))).Methods("POST")

	// GET /repo/{host}/{user}/{name}/{review}
	// Return a review from a repository
	api.HandleFunc("/repo/{host}/{user}/{name}/{review}", headers(getReviewHandler)).Methods("GET")

	// GET /api/repo/{host}/{user}/{name}
	// Get information and all the reviews on a repo
	api.HandleFunc("/repo/{host}/{user}/{name}", headers(getRepository)).Methods("GET")

	// POST /api/repo/{host}/{user}/{name}?access_token=xxx
	// Create a new link to github repository, return to that!
	api.HandleFunc("/repo/{host}/{user}/{name}", headers(restrict(newRepository))).Methods("POST")

	// GET /secret
	// A page to test secrecy!
	r.HandleFunc("/secret", headers(restrict(getSecret))).Methods("GET")

	// Serve ALL the static files!
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", r)

	fmt.Println("Loading http server on :8080...")

	http.ListenAndServe(":8080", nil)

}

func getSecret(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	fmt.Fprintln(w, "NCSS IS ILLUMINATTI")
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
	username, email, password := r.FormValue("username"), r.FormValue("email"), r.FormValue("password")
	uRe, eRe, pRe := usernameAndPasswordRegex.FindString(username), emailRegex.FindString(email), usernameAndPasswordRegex.FindString(username)

	if uRe == "" || eRe == "" || pRe == "" {
		fmt.Fprintln(w, "Username, password or email is not valid")
		return
	}

	c := session.DB(db).C("users")
	var u User

	if c.Find(bson.M{"username": username}).One(&u); u != (User{}) {
		fmt.Fprint(w, "That user already exists!")
		return
	}

	u = User{ID: bson.NewObjectId(), Username: username, Email: email, PasswordHash: password}

	if err := c.Insert(u); err != nil {
		fmt.Fprintln(w, "Unable to create that user at this time.")
		return
	}

	fmt.Fprintf(w, "%v", u)
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
	t.Claims["User"] = u.ID
	t.Claims["Expires"] = time.Now().Add(time.Minute * 15).Unix()

	tokenString, err := t.SignedString(signKey)

	if err != nil {
		fmt.Fprintln(w, "Error signing that token")
		return
	}

	json.NewEncoder(w).Encode(Token{Value: tokenString})
}

func getCurrentUserHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	id, ok := t.Claims["User"].(string)

	if !ok {
		fmt.Fprintln(w, "Could not cast interface to bson.ObjectId!")
		return
	}

	c := session.DB(db).C("users")
	var u User

	if c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&u); u == (User{}) {
		fmt.Fprintln(w, "COuld not find that user!")
		return
	}

	json.NewEncoder(w).Encode(u)
}

func newRepository(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]

	c := session.DB(db).C("repositories")
	var re Repository

	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re != (Repository{}) {
		fmt.Fprintln(w, "That repo already exist")
		return
	}

	re = Repository{ID: bson.NewObjectId(), Host: host, User: user, Name: name}

	if err := c.Insert(re); err != nil {
		fmt.Fprintln(w, "Currently unable to create that new repo")
		return
	}

	json.NewEncoder(w).Encode(re)
}

func newReviewHandler(w http.ResponseWriter, r *http.Request, t *jwt.Token) {
	vars := mux.Vars(r)
	host, user, name, review := vars["host"], vars["user"], vars["name"], r.FormValue("review")

	if review == "" {
		fmt.Fprintln(w, "Please let your review have some content?")
		return
	}

	c := session.DB(db).C("repositories")
	var rep Repository

	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&rep); rep == (Repository{}) {
		fmt.Fprintln(w, "a repo with that url doesnt exist...")
		return
	}

	c = session.DB(db).C("users")
	var u User

	if c.Find(bson.M{"_id": bson.ObjectIdHex(t.Claims["User"].(string))}).One(&u); u == (User{}) {
		fmt.Fprintln(w, "a user with that id doesnt exist...")
		return
	}

	c = session.DB(db).C("reviews")
	rev := Review{ID: bson.NewObjectId(), Content: review, From: u.ID, Repository: rep.ID}

	if err := c.Insert(rev); err != nil {
		fmt.Fprintln(w, "something went wrong while inserting the new review!")
		return
	}

	json.NewEncoder(w).Encode(rev)
}

func getReviewHandler(w http.ResponseWriter, r *http.Request) {

}

func getRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host, user, name := vars["host"], vars["user"], vars["name"]

	c := session.DB(db).C("repositories")
	var re Repository

	if c.Find(bson.M{"host": host, "user": user, "name": name}).One(&re); re == (Repository{}) {
		fmt.Fprintln(w, "that repo doesnt exist")
		return
	}

	json.NewEncoder(w).Encode(re)
}

func headers(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fn(w, r)
	}
}

func restrict(fn func(http.ResponseWriter, *http.Request, *jwt.Token)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.FormValue("access_token")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

		if err != nil {
			fmt.Fprintln(w, "That is not a valid token")
			fmt.Println(err)
			return
		}

		if !token.Valid {
			fmt.Fprintln(w, "Something obscurely strange happened to your token")
		}

		fn(w, r, token)
	}
}
