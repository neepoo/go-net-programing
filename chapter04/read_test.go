package chapter04

import (
	"crypto/rand"
	"errors"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	payload := make([]byte, 1<<24)                // 16Mb
	if _, err := rand.Read(payload); err != nil { // generate a random payload
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log("conn error", err)
			return
		}
		defer conn.Close()

		// write to client
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
		}
	}()

	// client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	allReceivedBytes := 0
	buf := make([]byte, 1<<19) // 512KB
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Error(err)
			}
			// EOF
			break
		}
		t.Logf("read %d bytes", n)
		allReceivedBytes += n
	}
	if allReceivedBytes != 1<<24 {
		t.Fatal("size is not equal")
	}
}
