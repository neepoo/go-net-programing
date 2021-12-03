package chapter05

import (
	"context"
	"fmt"
	"net"
	"time"
)

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to udp: %s: %w", addr, err)
	}

	go func() {
		go func() {
			// should be canceled
			<-ctx.Done()
			_ = s.Close()
		}()

		buf := make([]byte, 1024)
		for {
			// client to server
			n, clientAddr, err := s.ReadFrom(buf)
			//fmt.Println("clientAddr", clientAddr.String())
			if err != nil {
				fmt.Printf("server read error: %v\n", err)
				return
			}

			// server to client
			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				fmt.Printf("server write error: %v\n", err)
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}

func main() {
	now := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(time.Second))
	defer cancel()
	<-ctx.Done()
	// 1.000096943s
	fmt.Println(time.Now().Sub(now))
}
