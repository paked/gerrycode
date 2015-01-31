package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"regexp"
)

func generateKeys() {
	var err error

	signKey, err = ioutil.ReadFile(*privateKeyPath)

	if err != nil {
		fmt.Println("Could not find your private key!")
		panic(err)
	}

	verifyKey, err = ioutil.ReadFile(*publicKeyPath)

	if err != nil {
		fmt.Println("Could not find your public key!")
		panic(err)
	}

	signingMethod = jwt.GetSigningMethod("RS256")
}

func createUserRegex() {
	var err error
	usernameAndPasswordRegex, err = regexp.Compile(usernameAndPasswordRegexString)

	if err != nil {
		panic(err)
	}

	emailRegex, err = regexp.Compile(emailRegexString)

	if err != nil {
		panic(err)
	}
}
