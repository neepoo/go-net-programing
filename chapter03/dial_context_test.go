package chapter03

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	dl := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	var d net.Dialer
	d.Control = func(_, address string, _ syscall.RawConn) error {
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}

	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else if !nErr.Timeout() {
		t.Errorf("error is not a timeout : %v", err)
	}
	// 确保是五秒超时
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
