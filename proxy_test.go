package main

import (
	"net"
	"testing"
	"time"
)

var serverMap map[string]*net.UDPConn = make(map[string]*net.UDPConn)
var c config
var initialized bool = false

func setup_test(t *testing.T) {
	if initialized {
		return
	}

	initialized = true
	err := c.read("test")
	if err != nil {
		t.Error("unable to read config file", err)
	}

	setup(c.Nodes)
	makeServers(t)
	go startServer(c.UdpVersion, c.Port, c.Host)
}

func makeServers(t *testing.T) {
	for _, n := range c.Nodes {
		conn, err := makeConn(c.UdpVersion, n.Port, n.Host)
		if err != nil {
			t.Error("should be able to setup the servers", err)
		}
		serverMap[n.Name()] = conn
	}
}

func newConn(t *testing.T) (net.PacketConn, net.UDPAddr) {
	conn, err := net.ListenPacket(c.UdpVersion, "127.0.0.1:0")
	if err != nil {
		t.Error("should be able to create a connection", err)
	}
	addr := makeAddr(c.Port, "127.0.0.1")
	return conn, addr
}

func TestSetup(t *testing.T) {
	setup_test(t)
	name, err := cons.Get("statsd.metric.test")
	if err != nil {
		t.Error("cons should not return an error", err)
	}
	if name != "127.0.0.1:8129" {
		t.Error("expected name to be 127.0.0.1:8129, but it was", name)
	}
	name, err = cons.Get("statsd.metric.name")
	if err != nil {
		t.Error("cons should not return an error", err)
	}
	if name != "127.0.0.1:8127" {
		t.Error("expected name to be 127.0.0.1:8127, but it was", name)
	}
}

func TestOneMetric(t *testing.T) {
	setup_test(t)
	conn, addr := newConn(t)
	_, err := conn.WriteTo([]byte("statsd.metric.test:1|c"), &addr)
	if err != nil {
		t.Error("conn Write should not return an error", err)
	}

	readMetric("127.0.0.1:8129", "statsd.metric.test:1|c", t)
}

func TestMultipleMetrics(t *testing.T) {
	setup_test(t)
	conn, addr := newConn(t)
	_, err := conn.WriteTo([]byte("statsd.metric.test:1|c\nstatsd.metric.name:2|g"), &addr)
	if err != nil {
		t.Error("conn Write should not return an error", err)
	}

	readMetric("127.0.0.1:8129", "statsd.metric.test:1|c", t)
	readMetric("127.0.0.1:8127", "statsd.metric.name:2|g", t)
}

func readMetric(server string, metric string, t *testing.T) {
	node := serverMap[server]
	err := node.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		t.Error("unable to set node read deadline", err)
	}
	b := make([]byte, 1024)
	n, _, err := node.ReadFromUDP(b)
	if err != nil {
		t.Error("server Read should not return an error", err)
	}
	cmd := string(b[:n])
	if cmd != metric {
		t.Error("expected ", metric, " to be sent to this server, but received", cmd)
	}
}
