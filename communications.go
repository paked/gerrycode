package main

import (
	"encoding/json"
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

// NewComminicator returns a new communicator
func NewCommunicator(w http.ResponseWriter) *Communicator {
	return &Communicator{json.NewEncoder(w)}
}

// Communicator is an object used for sending JSON responses easily
type Communicator struct {
	e *json.Encoder
}

// Fail is an alias for FailWithData(msg, nil), used for user created errors
func (c Communicator) Fail(message string) {
	c.FailWithData(message, nil)
}

// FailWithData creates a Response object with the given message and data, and then sends it to the user
// It should be used for situations where the users input was the cause of a fault
func (c Communicator) FailWithData(message string, data interface{}) {
	r := c.message(message, NewFailedStatus(), data)
	c.writeMessage(r)
}

// OK is an alias for OKWithData(msg, nil), used for sending messages when everything is OK!
func (c Communicator) OK(message string) {
	c.OKWithData(message, nil)
}

// OKWithData creates a Response object with the given message and data, and then sends it to the user
// It should only be used when everything is OK
func (c Communicator) OKWithData(message string, data interface{}) {
	r := c.message(message, NewOKStatus(), data)
	c.writeMessage(r)
}

// Error is an alias for ErrorWithData(msg, nil), used for sending user created errors
func (c Communicator) Error(message string) {
	c.ErrorWithData(message, nil)
}

// ErrorWithData creates a Response object with the given message and data, and then sends it to the user
// It should be used when an error caused by the server or an external library is thrown
func (c Communicator) ErrorWithData(message string, data interface{}) {
	r := c.message(message, NewServerErrorStatus(), data)
	c.writeMessage(r)
}

// message creates a Response object
func (c Communicator) message(message string, status Status, data interface{}) Response {
	return Response{Message: message, Status: status, Data: data}
}

// writeMessage takes care of the actual sending of Response objects
func (c Communicator) writeMessage(r Response) {
	if err := c.e.Encode(r); err != nil {
		c.e.Encode("Something went very wrong!")
	}
}
