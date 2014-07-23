package main

import (
	"encoding/json"
	"log"
	"os"
)

type config struct {
	Nodes         []node
	Host          string
	Port          int
	UdpVersion    string
	CheckInterval int
}

func (c *config) read(env string) {
	file, _ := os.Open(env + ".json")
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		log.Fatal(err)
	}
}
