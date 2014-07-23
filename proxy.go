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
		n.Version = c.UdpVersion
		n.Addr = makeAddr(n.Port, n.Host)
		n.Add()
		clientMap[n.Name()] = n
	}
}

func startServer(c config) {
	conn, err := makeConn(c.UdpVersion, c.Port, c.Host)
	if err != nil {
		log.Fatal(err)
	}

	readPackets(conn)
}

func readPackets(conn *net.UDPConn) {
	defer conn.Close()
	for {
		b := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal(err)
		}
		p := packet{Length: n, Buffer: b, Conn: conn}
		go p.handle()
	}
}

func main() {
	var c config
	c.read("production")
	setup(c)
	go healthCheck(c.CheckInterval, c.UdpVersion)
	startServer(c)
}
