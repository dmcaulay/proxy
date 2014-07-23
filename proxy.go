package main

import (
	"log"
	"net"

	"github.com/stathat/consistent"
)

var clientMap map[string]*node = make(map[string]*node)
var cons = consistent.New()

func makeAddr(port int, host string) net.UDPAddr {
	return net.UDPAddr{Port: port, IP: net.ParseIP(host)}
}

func makeConn(version string, port int, host string) (*net.UDPConn, error) {
	addr := makeAddr(port, host)
	return net.ListenUDP(version, &addr)
}

func setup(c config) {
	// setup clients and hash ring
	cons.NumberOfReplicas = 1
	for i := 0; i < len(c.Nodes); i++ {
		n := &c.Nodes[i]
		n.Addr = makeAddr(n.Port, n.Host)
		n.Add()
		clientMap[n.Name()] = n
	}
}

func startServer(c config) error {
	conn, err := makeConn(c.UdpVersion, c.Port, c.Host)
	if err != nil {
		return err
	}

	return readPackets(conn)
}

func readPackets(conn *net.UDPConn) error {
	defer conn.Close()
	for {
		b := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			return err
		}
		p := packet{Length: n, Buffer: b, Conn: conn}
		go p.handle()
	}
}

func main() {
	var c config
	err := c.read("production")
	if err != nil {
		log.Fatal(err)
	}

	setup(c)
	go healthcheck(c.CheckInterval, c.UdpVersion)

	log.Fatal(startServer(c))
}
