package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"net/http"
)

func NewServer() *Server {
	s := &Server{}

	err := s.ConnectToDatabase("localhost")
	if err != nil {
		panic(err)
	}

	s.InitRouting()

	return s
}

type Server struct {
	Conn   *mgo.Session
	Router *mux.Router
}

func (s *Server) ConnectToDatabase(host string) error {
	var err error
	s.Conn, err = mgo.Dial(host)
	return err
}

func (s *Server) CloseConnectionDatabase() {
	s.Conn.Close()
}

func (s *Server) InitRouting() {
	s.Router = mux.NewRouter()
	api := s.Router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/user/create", Headers(NewUserHandler)).Methods("POST")

	api.HandleFunc("/user/login", Headers(LoginUserHandler)).Methods("POST")

	api.HandleFunc("/user", Headers(Restrict(GetCurrentUserHandler))).Methods("GET")

	api.HandleFunc("/user/{username}", Headers(GetUserHandler)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}/review", Headers(Restrict(NewReviewHandler))).Methods("POST")

	api.HandleFunc("/repo/{host}/{user}/{name}/{review}", Headers(GetReviewHandler)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}", Headers(GetRepository)).Methods("GET")

	api.HandleFunc("/repo/{host}/{user}/{name}", Headers(Restrict(NewRepository))).Methods("POST")

	s.Router.HandleFunc("/secret", Headers(Restrict(GetSecretHandler))).Methods("GET")

	// Serve ALL the static files!
	s.Router.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", s.Router)
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
			return
		}

		fn(w, r, token)
	}
}
