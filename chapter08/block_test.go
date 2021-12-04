package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

//func TestBlockIndefinitely(t *testing.T) {
//	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
//	_, _ = http.Get(ts.URL) //nolint:bodyclose
//	t.Fatal("client did not indefinitely block")
//}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	// tell client do per-request basic don't reuse tcp session
	//  tells Go’s HTTP client that it
	//should close the underlying TCP connection after reading the web server’s
	//response
	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}
	resp.Body.Close()
}
