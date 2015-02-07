// repo-review is an open source project for rating and comparing hosted git repositories.
// Written by Harrison Shoebridge (http://github.com/paked) available under the MIT license.
//
// Please contribute :) It makes me happy!
package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/paked/models"
)

var (
	server *Server

	privateKeyPath = flag.String("private", "keys/app.rsa", "path to the private key")
	publicKeyPath  = flag.String("public", "keys/app.rsa.pub", "path to the public key")
	db             = flag.String("db", "repo-reviews", "name of the database")
)

func init() {
	flag.Parse()

	generateKeys()
	createUserRegex()
}

func main() {
	models.Init("localhost", *db)
	server = NewServer()

	fmt.Println("Loading http server on :8080...")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
