package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/stathat/consistent"
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

	cons := consistent.New()
	cons.NumberOfReplicas = 1
	for _, node := range c.Nodes {
		name := fmt.Sprintf("%s:%d", node.Host, node.Port)
		fmt.Printf("add %s\n", name)
		cons.Add(name)
	}
}
