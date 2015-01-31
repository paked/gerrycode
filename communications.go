package main

import (
	"net/http"
)

// Response is used as a general response for JSON rest requests
type Response struct {
	Status  Status      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Status represents a general http error
type Status struct {
	Code    int    `bson:"code" json:"code"`
	Message string `bson:"message" json:"message"`
	Error   bool   `bson:"error" json:"error"`
}

// NewOKStatus returns a new Status object with no errors.
func NewOKStatus() Status {
	return Status{http.StatusOK, "Everything is awesome!", false}
}

// NewFailedStatus returns a new Status object saying a request failed.
func NewFailedStatus() Status {
	return Status{http.StatusConflict, "Well this is awkward...", true}
}

// NewForbiddenStatus returns a new Status object detailing a failure of authorization.
func NewForbiddenStatus() Status {
	return Status{http.StatusForbidden, "You can't go here :)", true}
}

// NewServerErrorStatus returns a new Status object saying a server error has occured
func NewServerErrorStatus() Status {
	return Status{http.StatusInternalServerError, "Something bad has happened, we're sending the calvalry.", true}
}
