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

func NewCommunicator(w http.ResponseWriter) *Communicator {
	return &Communicator{json.NewEncoder(w)}
}

type Communicator struct {
	*json.Encoder
}

func (c Communicator) Fail(message string) {
	c.FailWithData(message, nil)
}

func (c Communicator) FailWithData(message string, data interface{}) {
	r := c.message(message, NewFailedStatus(), data)
	c.writeMessage(r)
}

func (c Communicator) OK(message string) {
	c.OKWithData(message, nil)
}

func (c Communicator) OKWithData(message string, data interface{}) {
	r := c.message(message, NewOKStatus(), data)
	c.writeMessage(r)
}

func (c Communicator) Error(message string) {
	c.ErrorWithData(message, nil)
}

func (c Communicator) ErrorWithData(message string, data interface{}) {
	r := c.message(message, NewServerErrorStatus(), data)
	c.writeMessage(r)
}

func (c Communicator) message(message string, status Status, data interface{}) Response {
	return Response{Message: message, Status: status, Data: data}
}

func (c Communicator) writeMessage(r Response) {
	if err := c.Encode(r); err != nil {
		c.Encode("Something went very wrong!")
	}
}
