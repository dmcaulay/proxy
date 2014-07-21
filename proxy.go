package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/stathat/consistent"
)

type node struct {
	Host      string
	Port      int
	AdminPort int
}

type config struct {
	Nodes         []node
	Host          string
	Port          int
	UdpVersion    string
	CheckInterval int
}

type connMap map[string]*net.UDPConn

func makeAddr(port int, host string) net.UDPAddr {
	return net.UDPAddr{Port: port, IP: net.ParseIP(host)}
}

func makeConn(version string, port int, host string) *net.UDPConn {
	addr := makeAddr(port, host)
	conn, err := net.DialUDP(version, nil, &addr)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func main() {
	file, _ := os.Open("config.json")

	var c config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		log.Fatal(err)
	}

	var clientMap connMap = make(connMap)
	cons := consistent.New()
	cons.NumberOfReplicas = 1
	for _, n := range c.Nodes {
		name := fmt.Sprintf("%s:%d", n.Host, n.Port)
		fmt.Printf("add %s\n", name)
		cons.Add(name)
		clientMap[name] = makeConn(c.UdpVersion, n.Port, n.Host)
	}

	addr := makeAddr(c.Port, c.Host)
	conn, err := net.ListenUDP(c.UdpVersion, &addr)
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	var b [1024]byte
	for {
		n, _, err := conn.ReadFromUDP(b[:])
		if err != nil {
			log.Fatal(err)
		}
		buffer := bytes.NewBuffer(b[:])

		var pos int
		for {
			// read the next command
			line, err := buffer.ReadBytes('\n')
			if err != nil {
				log.Fatal(err)
			}

			// read the key
			metric, err := bytes.NewBuffer(line[:]).ReadBytes(':')
			if err != nil {
				log.Fatal(err)
			}
			key := string(metric[:])

			// get the client
			name, err := cons.Get(key)
			if err != nil {
				log.Fatal(err)
			}
			client, found := clientMap[name]
			if !found {
				log.Fatal("unknown client for key", key)
			}

			// write to the statsd server
			_, err = client.Write(line[:])
			if err != nil {
				log.Fatal(err)
			}

			// check position
			pos += len(line)
			if pos == n {
				break
			}
		}
	}
}
