package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	verifyKey, signKey []byte
	signingMethod      jwt.SigningMethod
)

// Server represents an instance of the go-github-review application.
type Server struct{}

// NewServer initializes a go-github-review server and then returns a pointer to it
func NewServer() *Server {
	s := &Server{}

	s.InitRouting()

	return s
}

// InitRouting creates all the necessary routes for go-github-review.
func (s *Server) InitRouting() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/user/create", s.headers(NewUserHandler)).Methods("POST")

	api.HandleFunc("/user/login", s.headers(LoginUserHandler)).Methods("POST")

	api.HandleFunc("/user", s.headers(s.restrict(GetCurrentUserHandler))).Methods("GET")

	api.HandleFunc("/user/{username}", s.headers(GetUserHandler)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}/review", s.headers(s.restrict(NewReviewHandler))).Methods("POST")

	api.HandleFunc("/repo/{host}/{user}/{name}/{review}", s.headers(GetReviewHandler)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}", s.headers(GetRepository)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}", s.headers(s.restrict(NewRepository))).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", r)
}

// headerify adds JSON headers onto a request.
func (s Server) headers(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fn(w, r)
	}
}

// s.restrict checks if a provided access_token is valid, if it is continue the request.
func (s Server) restrict(fn func(http.ResponseWriter, *http.Request, *jwt.Token)) http.HandlerFunc {
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
			return
		}

		fn(w, r, token)
	}
}
