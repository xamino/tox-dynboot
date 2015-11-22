package toxdynboot

import (
	"testing"
	"time"
)

func Test_ParseNodes(t *testing.T) {
	nodes, err := parseNodes()
	if err != nil {
		t.Fatal("ParseNodes:", err)
	}
	if len(nodes) == 0 {
		t.Error("Expected multiple nodes, got none.")
	}
}

func Test_IsAlive(t *testing.T) {
	// fake node
	node := &ToxNode{
		IPv4:      "0.1.2.3",
		IPv6:      "",
		Port:      12345,
		PublicKey: []byte("Invalid")}
	// test timeout
	f := func() chan bool {
		recv := make(chan bool, 1)
		recv <- isAlive(node, 100*time.Millisecond)
		return recv
	}
	select {
	case value := <-f():
		if value {
			t.Error("IsAlive returned valid node!")
		}
	case <-time.Tick(110 * time.Millisecond):
		t.Error("IsAlive took too long!")
	}
	// TODO more tests
}
