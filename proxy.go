package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/stathat/consistent"
)

type node struct {
	Host      string
	Port      int
	AdminPort int
	Version   string
	Conn      *net.UDPConn
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
	conn, err := makeConn(n.Version, n.Port, n.Host)
	if err == nil {
		n.Conn = conn
	}
}

func (n *node) Remove() {
	n.Conn = nil
	cons.Remove(n.Name())
}

type config struct {
	Nodes         []node
	Host          string
	Port          int
	UdpVersion    string
	CheckInterval int
}

type packet struct {
	Length int
	Buffer []byte
}

var clientMap map[string]node = make(map[string]node)
var cons = consistent.New()

func makeAddr(port int, host string) net.UDPAddr {
	return net.UDPAddr{Port: port, IP: net.ParseIP(host)}
}

func makeConn(version string, port int, host string) (*net.UDPConn, error) {
	addr := makeAddr(port, host)
	return net.DialUDP(version, nil, &addr)
}

func main() {
	// read config
	file, _ := os.Open("config.json")
	var c config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		log.Fatal(err)
	}

	// setup clients and hash ring
	cons.NumberOfReplicas = 1
	for _, n := range c.Nodes {
		fmt.Printf("add %s\n", n.Name())
		n.Version = c.UdpVersion
		clientMap[n.Name()] = n
		n.Add()
	}

	go healthCheck(c.CheckInterval)

	// start proxy server
	addr := makeAddr(c.Port, c.Host)
	conn, err := net.ListenUDP(c.UdpVersion, &addr)
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	// read packets
	for {
		var b []byte
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal(err)
		}

		if n == 0 {
			continue
		}

		go handlePacket(packet{Length: n, Buffer: b})
	}
}

func healthCheck(interval int) {
	healthMessage := []byte("health\r\n")
	up := []byte("up")
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	for {
		<-ticker.C
		for _, n := range clientMap {
			// connect to statsd admin port
			conn, err := makeConn(n.Version, n.AdminPort, n.Host)
			defer conn.Close()
			if err != nil {
				n.Remove()
				continue
			}

			// write health message
			_, err = conn.Write(healthMessage)
			if err != nil {
				n.Remove()
				continue
			}

			// read response
			var b []byte
			_, _, err = conn.ReadFromUDP(b)
			if err != nil {
				n.Remove()
				continue
			}

			// check to see if the node is up
			if bytes.Contains(b, up) {
				if n.Conn == nil {
					n.Add()
				}
			} else {
				n.Remove()
			}
		}
	}
}

func handlePacket(p packet) {
	buffer := bytes.NewBuffer(p.Buffer)
	var pos int

	for {
		// read the next command
		line, err := buffer.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		// read the key
		metric, err := bytes.NewBuffer(line).ReadBytes(':')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		key := string(metric)

		// get the client
		name, err := cons.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		n, found := clientMap[name]
		if !found {
			log.Fatal("unknown client for key", key)
		}

		conn := n.Conn
		if conn == nil {
			n.Remove()
		}

		// write to the statsd server
		_, err = conn.Write(line)
		if err != nil {
			n.Remove()
		}

		// check position
		pos += len(line)
		if pos == p.Length {
			break
		}
	}
}
