package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Conn      *net.UDPConn
	name      string
}

func (n *node) Name() string {
	if n.name == "" {
		n.name = fmt.Sprintf("%s:%d", n.Host, n.Port)
	}
	return n.name
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

type connMap map[string]node

func makeAddr(port int, host string) net.UDPAddr {
	return net.UDPAddr{Port: port, IP: net.ParseIP(host)}
}

func makeConn(version string, port int, host string) (*net.UDPConn, error) {
	addr := makeAddr(port, host)
	return net.DialUDP(version, nil, &addr)
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
		fmt.Printf("add %s\n", n.Name())
		clientMap[n.Name()] = n
		addNode(c.UdpVersion, n, cons)
	}

	go healthCheck(c.CheckInterval, c.UdpVersion, clientMap, cons)

	addr := makeAddr(c.Port, c.Host)
	conn, err := net.ListenUDP(c.UdpVersion, &addr)
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	for {
		var b []byte
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal(err)
		}

		go handlePacket(packet{Length: n, Buffer: b}, clientMap, cons)
	}
}

func addNode(version string, n node, cons *consistent.Consistent) {
	cons.Add(n.Name())
	conn, err := makeConn(version, n.Port, n.Host)
	if err == nil {
		n.Conn = conn
	}
}

func removeNode(n node, cons *consistent.Consistent) {
	n.Conn = nil
	cons.Remove(n.Name())
}

func healthCheck(interval int, version string, clientMap connMap, cons *consistent.Consistent) {
	healthMessage := []byte("health\r\n")
	up := []byte("up")
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	for {
		<-ticker.C
		for _, n := range clientMap {
			// connect to statsd admin port
			conn, err := makeConn(version, n.AdminPort, n.Host)
			defer conn.Close()
			if err != nil {
				removeNode(n, cons)
				continue
			}

			// write health message
			_, err = conn.Write(healthMessage)
			if err != nil {
				removeNode(n, cons)
				continue
			}

			// read response
			var b []byte
			_, _, err = conn.ReadFromUDP(b)
			if err != nil {
				removeNode(n, cons)
				continue
			}

			// check to see if the node is up
			if bytes.Contains(b, up) {
				if n.Conn == nil {
					addNode(version, n, cons)
				}
			} else {
				removeNode(n, cons)
			}
		}
	}
}

func handlePacket(p packet, clientMap connMap, cons *consistent.Consistent) {
	buffer := bytes.NewBuffer(p.Buffer)
	var pos int

	for {
		// read the next command
		line, err := buffer.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}

		// read the key
		metric, err := bytes.NewBuffer(line).ReadBytes(':')
		if err != nil {
			log.Fatal(err)
		}
		key := string(metric)

		// get the client
		name, err := cons.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		n, found := clientMap[name]
		if !found || n.Conn == nil {
			log.Fatal("unknown client for key", key)
		}

		// write to the statsd server
		_, err = n.Conn.Write(line)
		if err != nil {
			removeNode(n, cons)
		}

		// check position
		pos += len(line)
		if pos == p.Length {
			break
		}
	}
}
