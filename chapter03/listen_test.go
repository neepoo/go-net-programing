package chapter03

import (
	"io"
	"net"
	"testing"
)


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
