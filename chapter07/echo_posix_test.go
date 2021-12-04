//go:build darwin || linux
// +build darwin linux

package chapter07

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnixDatagram(t *testing.T) {
	dir, err := ioutil.TempDir("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	// server
	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid()))
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// client
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid()))
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	err = os.Chmod(cSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("ping")
	for i := 0; i < 3; i++ {
		_, err = client.WriteTo(msg, serverAddr)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1<<10)
	// read 3 "ping" message
	for i := 0; i < 3; i++ {
		n, addr, err := client.ReadFrom(buf)
		if err != nil {
			t.Fatal(err)
		}
		if addr.String() != serverAddr.String() {
			t.Fatalf("received reply from %q instead of %q",
				addr, serverAddr)
		}
		if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("expected reply %q; actual reply %q", msg,
				buf[:n])
		}
	}
}
