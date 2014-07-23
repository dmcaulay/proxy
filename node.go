package main

import (
	"fmt"
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
	cons.Add(n.Name())
}

func (n *node) Remove() {
	cons.Remove(n.Name())
}
