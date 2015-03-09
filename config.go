package main

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func passConfig(location string) (config, error) {
	var err error
	var c config

	bytes, err := ioutil.ReadFile(location)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
