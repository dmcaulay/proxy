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

var clientMap map[string]*node = make(map[string]*node)
var cons = consistent.New()

type config struct {
	Nodes         []node
	Host          string
	Port          int
	UdpVersion    string
	CheckInterval int
}

type node struct {
	Host      string
	Port      int
	AdminPort int
	Version   string
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

type packet struct {
	Length int
	Buffer []byte
	Conn   *net.UDPConn
}

func makeAddr(port int, host string) net.UDPAddr {
	return net.UDPAddr{Port: port, IP: net.ParseIP(host)}
}

func makeConn(version string, port int, host string) (*net.UDPConn, error) {
	addr := makeAddr(port, host)
	return net.ListenUDP(version, &addr)
}

func readConfig() config {
	file, _ := os.Open("config.json")
	var c config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		log.Fatal(err)
	}
	return c
}

func setup(c config) {
	// setup clients and hash ring
	cons.NumberOfReplicas = 1
	for i := 0; i < len(c.Nodes); i++ {
		n := &c.Nodes[i]
		n.Version = c.UdpVersion
		n.Addr = makeAddr(n.Port, n.Host)
		n.Add()
		clientMap[n.Name()] = n
	}
}

func startServer(c config) {
	conn, err := makeConn(c.UdpVersion, c.Port, c.Host)
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	// read packets
	for {
		b := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal(err)
		}
		go handlePacket(packet{Length: n, Buffer: b, Conn: conn})
	}
}

func handlePacket(p packet) {
	buffer := bytes.NewBuffer(p.Buffer[:p.Length])
	var pos int

	for {
		// read the next command
		line, err := buffer.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if len(line) == 0 {
			break
		}
		if err != io.EOF {
			line = line[:len(line)-1]
		}

		// read the key
		metric, err := bytes.NewBuffer(line).ReadBytes(':')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		key := string(metric[:len(metric)-1])

		// get the client
		name, err := cons.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		n, found := clientMap[name]
		if !found {
			log.Fatal("unknown client for key", key)
		}

		// write to the statsd server
		_, err = p.Conn.WriteToUDP(line, &n.Addr)
		if err != nil {
			n.Remove()
			continue
		}

		// check position
		pos += len(line)
		if pos == p.Length {
			break
		}
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
				n.Add()
			} else {
				n.Remove()
			}
		}
	}
}

func main() {
	c := readConfig()

	setup(c)

	go healthCheck(c.CheckInterval)

	startServer(c)
}
