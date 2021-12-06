package chapter11

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

func TestClientTls(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				u := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, u, http.StatusMovedPermanently)
				return
			}
		}),
	)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}

	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			/*
					It’s good practice to restrict your client’s curve
					preference to the P-256 curve
				 avoid the use of P-384 and P-521. P-256
				is immune to timing attacks, whereas P-384 and P-521 are not
			*/
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		},
	}
	//  You need to explicitly
	//bless your transport with HTTP/2 support
	err = http2.ConfigureTransport(tp)
	if err != nil {
		t.Fatal(err)
	}

	client2 := &http.Client{Transport: tp}

	_, err = client2.Get(ts.URL)
	/*
		The first call
		to the test server results in an error because your client doesn’t trust the
		server certificate’s signatory.
	*/
	if err == nil || !strings.Contains(err.Error(), "certificate signed by unknown authority") {
		t.Fatalf("expected unknown authority error; actual: %q", err)
	}

	tp.TLSClientConfig.InsecureSkipVerify = true
	resp, err = client2.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
}

func TestClientTLSGoogle(t *testing.T) {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 30 * time.Second},
		"tcp",
		"www.baidu.com:443",
		&tls.Config{
			MinVersion:       tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{tls.CurveP256},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	//返回tls连接的基本信息
	state := conn.ConnectionState()
	t.Logf("TLS 1.%d", state.Version-tls.VersionTLS10)
	t.Log(tls.CipherSuiteName(state.CipherSuite))
	t.Log(state.VerifiedChains[0][0].Issuer.Organization[0])
	conn.Close()

}
