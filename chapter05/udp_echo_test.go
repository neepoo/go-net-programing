package chapter05

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestEchoServerUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	t.Log(serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	// best practice
	defer cancel()

	// NOTICE: this isn't Dial
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	msg := []byte("ping")
	_, err = client.WriteTo(msg, serverAddr)
	if err != nil {
		t.Fatalf("client write to server error, %v", err)
	}

	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatalf("client read from server error: %v", err)
	}
	if addr.String() != serverAddr.String() {
		t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
	}
	if !bytes.Equal(msg, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
	}
	//time.Sleep(10 * time.Second)
}
