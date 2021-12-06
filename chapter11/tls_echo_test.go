package chapter11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

const CertFile = "localhost.pem"
const KeyFile = "localhost-key.pem"

func TestEchoServerTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverAddress := "localhost:34443"
	maxIdle := time.Second
	server := NewTLSServer(ctx, serverAddress, maxIdle, nil)
	done := make(chan struct{})

	go func() {

		err := server.ListenAndServerTLS(CertFile, KeyFile)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()
	// 等待服务器设置并启动完成
	server.Ready()

	// 明确信任服务器的证书
	cert, err := ioutil.ReadFile(CertFile)
	if err != nil {
		t.Fatal(err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		t.Fatal("failed to append certificate to pool")
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		// 客户端使用这些cert去认证服务器
		// cert pin
		// You pass tls.Dial the tls.Config with the pinned server certificate 1
		RootCAs: certPool,
		// 不推荐
		//InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", serverAddress, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	hello := []byte("Hello TLS!")
	_, err = conn.Write(hello)
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 1<<10)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	if actual := b[:n]; !bytes.Equal(hello, actual) {
		t.Fatalf("expected %q; actual %q", hello, actual)
	}
	time.Sleep(2 * maxIdle)
	_, err = conn.Read(b)
	if !errors.Is(err, io.EOF) {
		//  showing the server closed the socket.
		t.Fatal(err)
	}
}
