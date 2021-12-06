package chapter11

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type Server struct {
	ctx   context.Context
	ready chan struct{}

	addr      string
	maxIdle   time.Duration
	tlsConfig *tls.Config
}

func NewTLSServer(ctx context.Context, address string, maxIdle time.Duration, tlsConfig *tls.Config) *Server {
	return &Server{
		ctx:   ctx,
		ready: make(chan struct{}),
		addr:  address,
		// 针对的是conn对象,就是tcp会话
		maxIdle:   maxIdle,
		tlsConfig: tlsConfig,
	}
}

func (s *Server) Ready() {
	if s.ready != nil {
		<-s.ready
	}

}

// ListenAndServerTLS method accepts full paths to a certificate and a
// private key and returns an error.
func (s *Server) ListenAndServerTLS(certFn, keyFn string) error {
	if s.addr == "" {
		s.addr = "localhost:443"
	}
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("binding to tcp %s: %w", s.addr, err)
	}

	if s.ctx != nil {
		go func() {
			<-s.ctx.Done()
			l.Close()
		}()
	}
	return s.ServeTLS(l, certFn, keyFn)
}

func (s Server) ServeTLS(l net.Listener, certFn, keyFn string) error {
	if s.tlsConfig == nil {
		s.tlsConfig = &tls.Config{
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
			// noticePreferServerCipherSuites is meaningful to the server only, and it makes the
			//server use its preferred cipher suite instead of deferring to the client’s
			//preference.
			PreferServerCipherSuites: true,
		}
	}

	if len(s.tlsConfig.Certificates) == 0 && s.tlsConfig.GetCertificate == nil {
		// load cert
		cert, err := tls.LoadX509KeyPair(certFn, keyFn)
		if err != nil {
			return fmt.Errorf("loading key pair: %v", err)
		}
		s.tlsConfig.Certificates = []tls.Certificate{cert}
	}
	// tls over tcp
	tlsListener := tls.NewListener(l, s.tlsConfig)
	if s.ready != nil {
		close(s.ready)
	}

	for {
		// wait until connection coming
		// Since you’re using a TLS-aware listener, it returns connection objects with
		//underlying TLS support
		conn, err := tlsListener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %v", err)
		}

		go func() {
			defer conn.Close()

			for {
				if s.maxIdle > 0 {
					err := conn.SetDeadline(time.Now().Add(s.maxIdle))
					if err != nil {
						return
					}
				}

				// read request
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					return
				}

				_, err = conn.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}()

	}
}
