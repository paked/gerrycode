package main

import (
	"github.com/paked/gerrycode/communicator"
	"net/http"
)

func NewCommunicator(w http.ResponseWriter) *communicator.Communicator {
	return communicator.New(w)
}
