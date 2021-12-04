package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
	resp, err := http.Get("https://www.time.gov/")
	if err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	io.Copy(buf, resp.Body)
	t.Logf("%v\n", buf.String())
	_ = resp.Body.Close()

	now := time.Now().Round(time.Second)
	date := resp.Header.Get("Date")
	if date == "" {
		t.Fatal("no Date header received from time.gov")
	}
	dt, err := time.Parse(time.RFC1123, date)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("time.gov: %s (skew %s)", dt, now.Sub(dt))
}
