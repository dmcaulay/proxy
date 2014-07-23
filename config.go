package main

import (
	"encoding/json"
	"os"
)

type config struct {
	Nodes         []node
	Host          string
	Port          int
	UdpVersion    string
	CheckInterval int
}

func (c *config) read(env string) error {
	file, err := os.Open(env + ".json")
	if err != nil {
		return err
	}
	return json.NewDecoder(file).Decode(&c)
}
