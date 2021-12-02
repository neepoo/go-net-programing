package chapter04

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
)

// more generic proxy
func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)

	if toIsReader && fromIsWriter {
		// replies
		go func() {
			_, err := io.Copy(fromWriter, toReader)
			if errors.Is(err, io.EOF) {
				fmt.Println("proxy replies EOF!!!")
			}
		}()
	}
	_, err := io.Copy(to, from)
	return err
}

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup

	// server listens for a "ping" message and responds with a
	// "pong" message. All other messages are echoed back to the client.
	server, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			// handle
			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							t.Error(err)
						}
						return
					}
					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong"))
					default:
						_, err = c.Write(buf[:n])
					}
					if err != nil {
						if !errors.Is(err, io.EOF) {
							t.Error(err)
						}
						return
					}
				}

			}(conn)

		}
	}()

	// proxyServer proxies messages from client connections to the
	// destinationServer. Replies from the destinationServer are proxied
	// back to the clients.
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()
				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				defer to.Close()

				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}

			}(conn)

		}
	}()

	client, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	msgs := []struct{ Msg, Replay string }{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	for i, m := range msgs {
		_, err = client.Write([]byte(m.Msg))
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 1024)

		n, err := client.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf[:n])
		t.Logf("%q -> proxy -> %q", m.Msg, actual)
		if actual != m.Replay {
			t.Errorf("%d: expected replay: %q; actual :%q", i, m.Replay, actual)
		}
	}
	_ = client.Close()
	_ = proxyServer.Close()
	_ = server.Close()
	wg.Wait()
}
