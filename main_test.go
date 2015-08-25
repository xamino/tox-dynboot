package toxdynboot

import (
	"testing"
	"time"
)

func Test_IsAlive(t *testing.T) {
	// fake node
	node := &ToxNode{
		IPv4:      "0.1.2.3",
		IPv6:      "",
		Port:      12345,
		PublicKey: []byte("Invalid"),
		Name:      "invalid"}
	// test timeout
	f := func() chan bool {
		recv := make(chan bool, 1)
		recv <- IsAlive(node, 100*time.Millisecond)
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
