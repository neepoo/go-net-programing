package chapter07

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnix(t *testing.T) {
	dir, err := ioutil.TempDir("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	// make sure everyone has read and write access to the socket
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	// client send fin
	defer conn.Close()
	i2 := 8 << 10
	msg := make([]byte, i2)
	i3, err := rand.Read(msg)
	t.Logf("size:%d %d\n", i2, i3)
	if err != nil {
		t.Errorf("generate random payload error: %v\n", err)
	}

	for i := 0; i < 3; i++ {
		n, err := conn.Write(msg)
		if err != nil || n != i2 {
			t.Fatalf("send byte:%d, expected:%d, err:=%v\n", n, 8<<20, err)
		}
	}
	buf := make([]byte, 24<<10)
	// read once from the server
	readSize := 0
	for {
		sz, err := conn.Read(buf[readSize:])
		if sz == 0 {
			break
		}
		readSize += sz
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.Log("EOF")
				break
			}
			t.Fatal(err)
		}
	}
	t.Logf("read from server %d bytes, expected:%d\n", readSize, 24<<10)
	want := bytes.Repeat(msg, 3)
	if !bytes.Equal(want, buf[:readSize]) {
		t.Fatalf("expected reply %q; actual reply %q", want,
			buf[:readSize])
	}
}
