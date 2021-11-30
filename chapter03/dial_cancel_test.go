package chapter03

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

// 按照需求cancel
func TestDialContextCancel(t *testing.T) {
	n := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})

	go func() {
		defer func() { sync <- struct{}{} }()

		var d net.Dialer

		d.Control = func(_, address string, _ syscall.RawConn) error {
			time.Sleep(time.Second)
			return nil
		}
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}
		conn.Close()
		t.Error("connection did not time out")
	}()

	//This causes the DialContext method to immediately
	//return with a non- nil error, exiting the goroutine
	cancel()
	<-sync

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("expected canceled context; actual: %q", ctx.Err())
	}
	t.Log(time.Since(n))
}
