package chapter05

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

// You can establish a UDP connection that implements the net.Conn interface
// so that your code behaves indistinguishably from a TCP net.Conn
// Using net.Conn with your UDP-based connections can
// prevent interlopers from sending you messages and eliminate the need to
// check the sender’s address on every reply you receive.

func TestDialUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 确保关闭 "udp server"
	defer cancel()

	client, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// 中断client
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	interrupt := []byte("pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	interloper.Close()
	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	ping := []byte("ping")
	_, err = client.Write(ping)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 1<<10)
	n, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	// check echo function
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}

	err = client.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("unexpected packet")
	}
}
