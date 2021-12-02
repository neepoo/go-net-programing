package chapter04

import (
	"io"
	"log"
	"net"
	"os"
	"testing"
)

type Monitor struct {
	*log.Logger
}

func (m *Monitor) Write(p []byte) (n int, err error) {
	return len(p), m.Output(2, string(p))
}

func TestExampleMonitor(t *testing.T) {
	monitor := &Monitor{log.New(os.Stdout, "monitor: ", log.LstdFlags)}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}
	done := make(chan struct{})

	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		b := make([]byte, 1<<10)
		r := io.TeeReader(conn, monitor)
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		w := io.MultiWriter(conn, monitor)
		_, err = w.Write(b[:n])
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

	}()

	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}
	_, err = client.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}
	_ = client.Close()
	v, ok := <-done
	monitor.Println(v, ok, "received from done")
}
