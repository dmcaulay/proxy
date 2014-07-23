package main

import (
	"bytes"
	"log"
	"time"
)

func healthcheck(interval int, version string) {
	healthMessage := []byte("health\r\n")
	up := []byte("up")
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	conn, err := makeConn(version, 0, "0.0.0.0")
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	for {
		<-ticker.C
		for _, n := range clientMap {
			// write health message
			_, err = conn.WriteToUDP(healthMessage, &n.Addr)
			if err != nil {
				n.Remove()
				continue
			}

			// read response
			b := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(time.Duration(100) * time.Millisecond))
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
