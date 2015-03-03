package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/dgrijalva/jwt-go"
)

func generateKeys() {
	var err error

	signKey, err = ioutil.ReadFile(*privateKeyPath)

	if err != nil {
		fmt.Println("Could not find your private key!")
		os.Exit(1)
	}

	verifyKey, err = ioutil.ReadFile(*publicKeyPath)

	if err != nil {
		fmt.Println("Could not find your public key!")
		os.Exit(1)
	}

	signingMethod = jwt.GetSigningMethod("RS256")
}

func createUserRegex() {
	credentialsRegex = regexp.MustCompile(credentialsRaw)
	emailRegex = regexp.MustCompile(emailRaw)
}
