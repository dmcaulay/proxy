package main

import (
	"flag"
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

func setup(nodes []node) {
	// setup clients and hash ring
	cons.NumberOfReplicas = 1
	for i := 0; i < len(nodes); i++ {
		n := &nodes[i]
		n.Addr = makeAddr(n.Port, n.Host)
		n.Add()
		clientMap[n.Name()] = n
	}
}

func startServer(version string, port int, host string) error {
	conn, err := makeConn(version, port, host)
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
	env := flag.String("e", "development", "the program environment")
	flag.Parse()

	var c config
	err := c.read(*env)
	if err != nil {
		log.Fatal(err)
	}

	setup(c.Nodes)
	go healthcheck(c.CheckInterval, c.UdpVersion, c.Nodes)

	log.Fatal(startServer(c.UdpVersion, c.Port, c.Host))
}
