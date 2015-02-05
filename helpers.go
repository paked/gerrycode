package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
	"regexp"
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
	var err error
	usernameAndPasswordRegex, err = regexp.Compile(usernameAndPasswordRegexString)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	emailRegex, err = regexp.Compile(emailRegexString)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
