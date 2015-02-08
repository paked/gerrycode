// repo-review is an open source project for rating and comparing hosted git repositories.
// Written by Harrison Shoebridge (http://github.com/paked) available under the MIT license.
//
// Please contribute :) It makes me happy!
package main

import (
	"flag"
	"fmt"

	"github.com/paked/models"
)

var (
	server *Server

	privateKeyPath = flag.String("private", "keys/app.rsa", "path to the private key")
	publicKeyPath  = flag.String("public", "keys/app.rsa.pub", "path to the public key")
	db             = flag.String("db", "repo-reviews", "name of the database")

	host = flag.String("host", "localhost", "host to start the server on")
	port = flag.String("port", "8080", "port to listen on")
)

func init() {
	flag.Parse()

	generateKeys()
	createUserRegex()
}

func main() {
	models.Init("localhost", *db)
	server = NewServer()

	fmt.Println(server.Run(*host, *port))
}
