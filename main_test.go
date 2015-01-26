package main

import (
	"regexp"
	"testing"
)

func TestValidUsername(t *testing.T) {
	r, err := regexp.Compile(`^[a-zA-Z]\w*[a-zA-Z]$`)
	if err != nil {
		t.Fatal("Not valid regex.")
	}

	if s := r.FindString("bob"); s == "" {
		t.Fatal("bob is a valid username")
	}

	if s := r.FindString("ab"); s == "" {
		t.Fatal("ab is a valid username")
	}

	if s := r.FindString("Abc"); s == "" {
		t.Fatal("Abc is a valid username")
	}

	if s := r.FindString("_a"); s != "" {
		t.Fatal("A username cannot start with an underscore")
	}

	if s := r.FindString("0a"); s != "" {
		t.Fatal("A username cannot start with a number")
	}

	if s := r.FindString("a||||||$%&b"); s != "" {
		t.Fatal("A username cannot have non alphanumerical characters.")
	}
}

func TestValidEmail(t *testing.T) {
	r, err := regexp.Compile(`^.*\@.*$`)

	if err != nil {
		t.Fatal("Not valid regex.")
	}

	if s := r.FindString("harrison@theshoebridges.com"); s == "" {
		t.Fatal("That email contains an @")
	}

	if s := r.FindString("har(at)lolololol.com"); s != "" {
		t.Fatal("That email does not contain an @")
	}
}
