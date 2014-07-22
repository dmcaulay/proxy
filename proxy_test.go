package main

import "testing"

func TestConsistent(t *testing.T) {
	c := readConfig()
	setup(c)
	name, err := cons.Get("statsd.metric.name")
	if err != nil {
		t.Error("cons should not return an error")
	}
	if name == "" {
		t.Error("name should be initialized")
	}
}
