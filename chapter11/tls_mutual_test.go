package chapter11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"strings"
	"testing"
)

func caCertPool(caCertFn string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(caCertFn)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	/*
		The certificate pool serves as a source of trusted certificates. The client puts
		the server’s certificate in its certificate pool, and vice versa.
	*/
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("failed to add certificate to pool")
	}
	return certPool, nil
}

func TestMutualTLSAuthentication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Before creating the server, you need to first populate a new CA certifi-
	//cate pool with the client’s certificate
	serverPool, err := caCertPool("clientCert.pem")
	if err != nil {
		t.Fatal(err)
	}

	cert, err := tls.LoadX509KeyPair("serverCert.pem", "serverKey.pem")
	if err != nil {
		t.Fatalf("loading key pair: %v", err)
	}
	serverConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			return &tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				//ClientAuth:   tls.RequireAndVerifyClientCert | tls.RequestClientCert,
				// root ca to auth client
				ClientCAs:                serverPool,
				CurvePreferences:         []tls.CurveID{tls.CurveP256},
				MinVersion:               tls.VersionTLS13,
				PreferServerCipherSuites: true,

				// implements the verification process that authenticates the cli-
				//ent by using its IP address and its certificate’s common name and alterna-
				//tive names.

				//  The server calls this method
				//after the normal certificate verification checks

				// The leaf certificate is the last certificate in the certificate chain given to
				//the server by the client. The leaf certificate contains the client’s public key.
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					opts := x509.VerifyOptions{
						//Create a new x509.VerifyOptions object and modify the KeyUsages method
						//to indicate you want to perform client authentication
						KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
						// The server uses this pool as its trusted
						//certificate source during verification.
						Roots: serverPool,
					}

					ip := strings.Split(hello.Conn.RemoteAddr().String(), ":")[0]
					hostnames, err := net.LookupAddr(ip)

					// You’ll find each leaf certificate at index 0 in each
					//verifiedChains slice. In other words, you can find the leaf certificate of the
					//first chain at verifiedChains[0][0]
					for _, chain := range verifiedChains {
						opts.Intermediates = x509.NewCertPool()
						for _, cert := range chain[1:] {
							opts.Intermediates.AddCert(cert)
						}

						for _, hostname := range hostnames {
							opts.DNSName = hostname
							_, err = chain[0].Verify(opts)
							if err == nil {
								return nil
							}
						}
					}
					return errors.New("client authentication error")

				},
			}, nil
		},
	}

	serverAddress := "localhost:44443"
	server := NewTLSServer(ctx, serverAddress, 0, serverConfig)
	done := make(chan struct{})

	go func() {
		err := server.ListenAndServerTLS("serverCert.pem", "serverKey.pem")
		if err != nil && !strings.Contains(err.Error(),
			"use of closed network connection") {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()

	server.Ready()

	clientPool, err := caCertPool("serverCert.pem")
	if err != nil {
		t.Fatal(err)
	}
	clientCert, err := tls.LoadX509KeyPair("clientCert.pem", "clientKey.pem")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := tls.Dial("tcp", serverAddress, &tls.Config{
		Certificates:     []tls.Certificate{clientCert},
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS13,
		/*
			The client then uses the certificate pool in the RootCAs field
			of its TLS configuration 3, meaning the client will trust only server cer-
			tificates signed by serverCert.pem
		*/
		RootCAs: clientPool,
	})

	/*
		It’s worth noting that the client and server have not initialized a TLS
		session yet. They haven’t completed the TLS handshake. If tls.Dial returns
		an error, it isn’t because of an authentication issue but more likely a TCP
		connection issue.
	*/
	if err != nil {
		t.Fatal(err)
	}

	hello := []byte("hello mutual tls")
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
	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	<-done
}
