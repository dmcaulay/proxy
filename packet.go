package main

import (
	"bytes"
	"io"
	"log"
	"net"
)

type packet struct {
	Length int
	Buffer []byte
}

func (p *packet) handle(send func([]byte, *net.UDPAddr) error) {
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
		err = send(line, &n.Addr)
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
