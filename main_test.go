package main

import (
	"regexp"
	"testing"
)

// TestValidUsername tests if the username regex works as intended.
func TestValidUsername(t *testing.T) {
	r, err := regexp.Compile(usernameAndPasswordRegexString)

	if err != nil {
		t.Error("Not valid regex.")
	}

	if s := r.FindString("bob"); s == "" {
		t.Error("bob is a valid username")
	}

	if s := r.FindString("ab"); s == "" {
		t.Error("ab is a valid username")
	}

	if s := r.FindString("Abc"); s == "" {
		t.Error("Abc is a valid username")
	}

	if s := r.FindString("_a"); s != "" {
		t.Error("A username cannot start with an underscore")
	}

	if s := r.FindString("0a"); s != "" {
		t.Error("A username cannot start with a number")
	}

	if s := r.FindString("a||||||$%&b"); s != "" {
		t.Error("A username cannot have non alphanumerical characters.")
	}
}

// TestValidEmail tests if the email regex works as intended.
func TestValidEmail(t *testing.T) {
	r, err := regexp.Compile(emailRegexString)

	if err != nil {
		t.Error("Not valid regex.")
	}

	if s := r.FindString("harrison@theshoebridges.com"); s == "" {
		t.Error("That email contains an @")
	}

	if s := r.FindString("har(at)lolololol.com"); s != "" {
		t.Error("That email does not contain an @")
	}
}

func TestCreateUser(t *testing.T) {

}

func TestCreateUserFail(t *testing.T) {

}

func TestUserDelete(t *testing.T) {

}

func TestRepositoryCreate(t *testing.T) {

}

func TestRepositoryCreateFail(t *testing.T) {

}

func TestAddReview(t *testing.T) {

}

func TestAddReviewFail(t *testing.T) {

}

func TestDeleteReview(t *testing.T) {

}

func TestDeleteReviewFail(t *testing.T) {

}

func TestEditReview(t *testing.T)
