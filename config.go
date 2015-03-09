package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func passConfig(location string) (Config, error) {
	var err error
	var c Config

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
