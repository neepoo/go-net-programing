package chapter03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		// after create connection before dial
		Control: func(_, address string, _ syscall.RawConn) error {
			//  mocking a DNS time-out error.
			return &net.DNSError{
				Err:         "connection time out",
				Name:        address,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		Timeout: timeout,
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	c, err := DialTimeout("tcp", "10.0.0.0:80", 5 *time.Second)
	if err == nil{
		c.Close()
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok{
		t.Fatal(err)
	}
	if !nErr.Timeout(){
		t.Fatal("error is not a timeout")
	}
}