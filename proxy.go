package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Node struct {
	Host      string
	Port      int
	AdminPort int
}

type Config struct {
	Nodes         []Node
	UdpVersion    string
	Host          string
	Port          int
	CheckInterval int
	CacheSize     int
}

func main() {
	file, _ := os.Open("config.json")

	var c Config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", c)
}
