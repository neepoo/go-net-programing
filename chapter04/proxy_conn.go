package chapter04

import (
	"io"
	"net"
)

func proxyConn(src, dst string) error {
	connSource, err := net.Dial("tcp", src)
	if err != nil {
		return err
	}
	defer connSource.Close()

	connDst, err := net.Dial("tcp", dst)
	if err != nil {
		return err
	}
	defer connDst.Close()

	go func() {
		// handle replies
		/*		You don’t need to worry about leaking this goroutine, since io.Copy will
				return when either connection is closed.
		*/io.Copy(connSource, connDst)
	}()
	_, err = io.Copy(connDst, connSource)
	return err
}
