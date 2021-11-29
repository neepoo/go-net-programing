package chapter03

import (
	"io"
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		listener.Close()
	}()
	t.Logf("bound to %q", listener.Addr())
	for {
		/*
			 This method will block until the listener
			detects an incoming connection and completes the TCP handshake process
			between the client and the server.
		*/
		conn, err := listener.Accept()
		if err != nil {
			t.Fatalf("accept error: %v\n", err)
		}
		go func(c net.Conn) {
			defer c.Close()
			t.Logf("remote addr:%v\n", conn.RemoteAddr())
		}(conn)
	}
}

func TestDial(t *testing.T) {
	lister, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})

	// server
	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			conn, err := lister.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					// send fin
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}else {
							t.Log("EOF")
						}
						// exit
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}
	}()

	// client
	conn, err := net.Dial("tcp", lister.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	<-done
	lister.Close()
	<-done
}
