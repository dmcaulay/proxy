package main

import (
	"log"
	"net"
	"testing"
)

var serverMap map[string]*net.UDPConn = make(map[string]*net.UDPConn)
var c config
var initialized bool = false

func setup_test() {
	if initialized {
		return
	}
	initialized = true
	c.read("test")
	setup(c)
	makeServers()
	go startServer(c)
}

func makeServers() {
	for _, n := range c.Nodes {
		conn, err := makeConn(c.UdpVersion, n.Port, n.Host)
		if err != nil {
			log.Fatal(err)
		}
		serverMap[n.Name()] = conn
	}
}

func serverConn(t *testing.T) (net.PacketConn, net.UDPAddr) {
	conn, err := net.ListenPacket(c.UdpVersion, "127.0.0.1:0")
	if err != nil {
		t.Error("should be able to create a connection")
	}
	addr := makeAddr(c.Port, "127.0.0.1")
	return conn, addr
}

func TestConsistent(t *testing.T) {
	setup_test()
	name, err := cons.Get("statsd.metric.test")
	if err != nil {
		t.Error("cons should not return an error")
	}
	if name != "127.0.0.1:8129" {
		t.Error("expected name to be 127.0.0.1:8129, but it was", name)
	}
	name, err = cons.Get("statsd.metric.name")
	if err != nil {
		t.Error("cons should not return an error")
	}
	if name != "127.0.0.1:8131" {
		t.Error("expected name to be 127.0.0.1:8131, but it was", name)
	}
}

func TestOneMetric(t *testing.T) {
	setup_test()
	conn, addr := serverConn(t)
	_, err := conn.WriteTo([]byte("statsd.metric.test:1|c"), &addr)
	if err != nil {
		t.Error("conn Write should not return an error")
	}

	readMetric("127.0.0.1:8129", "statsd.metric.test:1|c", t)
}

func TestMultipleMetrics(t *testing.T) {
	setup_test()
	conn, addr := serverConn(t)
	_, err := conn.WriteTo([]byte("statsd.metric.test:1|c\nstatsd.metric.name:2|g"), &addr)
	if err != nil {
		t.Error("conn Write should not return an error")
	}

	readMetric("127.0.0.1:8129", "statsd.metric.test:1|c", t)
	readMetric("127.0.0.1:8131", "statsd.metric.name:2|g", t)
}

func readMetric(server string, metric string, t *testing.T) {
	node := serverMap[server]
	b := make([]byte, 1024)
	n, _, err := node.ReadFromUDP(b)
	if err != nil {
		t.Error("server Read should not return an error")
	}
	cmd := string(b[:n])
	if cmd != metric {
		t.Error("expected ", metric, " to be sent to this server, but received", cmd)
	}
}
