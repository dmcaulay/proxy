package main

import (
	"fmt"
	"log"
	"net"
)

type node struct {
	Host      string
	Port      int
	AdminPort int
	Addr      net.UDPAddr
	name      string
}

func (n *node) Name() string {
	if n.name == "" {
		n.name = fmt.Sprintf("%s:%d", n.Host, n.Port)
	}
	return n.name
}

func (n *node) Add() {
	log.Println("adding node", n.Name())
	cons.Add(n.Name())
}

func (n *node) Remove() {
	log.Println("removing node", n.Name())
	cons.Remove(n.Name())
}
