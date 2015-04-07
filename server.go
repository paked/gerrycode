package main

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var (
	verifyKey, signKey []byte
	signingMethod      jwt.SigningMethod
)

// Server represents an instance of the gerrycode application.
type Server struct{}

// NewServer initializes a gerrycode server and then returns a pointer to it
func NewServer() *Server {
	s := &Server{}

	s.InitRouting()

	return s
}

// InitRouting creates all the necessary routes for gerrycode
func (s *Server) InitRouting() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/user/register", s.headers(NewUserHandler)).Methods("POST")

	api.HandleFunc("/user/login", s.headers(LoginUserHandler)).Methods("POST")

	api.HandleFunc("/user", s.headers(s.restrict(GetCurrentUserHandler))).Methods("GET")

	api.HandleFunc("/user/projects", s.headers(s.restrict(GetUsersProjectsHandler))).Methods("GET")

	api.HandleFunc("/user/{username}", s.headers(GetUserHandler)).Methods("GET")

	api.HandleFunc("/user/git/repositories", s.headers(s.restrict(GetUsersRepositories))).Methods("GET")

	api.HandleFunc("/project/new", s.headers(s.restrict(PostCreateProjectHandler))).Methods("POST")

	api.HandleFunc("/project/{id}", s.headers(GetProjectHandler)).Methods("GET")

	api.HandleFunc("/project/{id}/update", s.headers(s.restrict(PostEditProject))).Methods("POST")

	api.HandleFunc("/project/{id}/flags/new", s.headers(s.restrict(PostFlagForFeedbackHandler))).Methods("POST")

	api.HandleFunc("/project/{id}/flags", s.headers(GetProjectsFlagsHandler)).Methods("GET")

	api.HandleFunc("/project/{id}/flags/{flag}", s.headers(GetFlagHandler)).Methods("GET")

	api.HandleFunc("/project/{id}/flags/{flag}/feedback/new", s.headers(s.restrict(PostFeedbackOnFlagHandler))).Methods("POST")

	api.HandleFunc("/project/{id}/flags/{flag}/feedback", s.headers(GetAllFeedbackForFlagHandler)).Methods("GET")

	api.HandleFunc("/top/projects", s.headers(GetTopProjects)).Methods("GET")

	api.HandleFunc("/reg", s.restrict(PostLinkUserAccount)).Methods("GET")

	api.HandleFunc("/oauth", GetAuthedGithubAccount).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", r)
}

// Run runs the http server.
func (s Server) Run(host, port string) error {
	address := fmt.Sprint(host, ":", port)
	fmt.Println("Starting server on", address)
	return http.ListenAndServe(address, nil)
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
		c := NewCommunicator(w)
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

		if err != nil {
			c.Fail("You are not using a valid token:" + err.Error())
			fmt.Println(err)
			return
		}

		if !token.Valid {
			c.Fail("Something obscurely weird happened to your token!")
			return
		}

		fn(w, r, token)
	}
}
