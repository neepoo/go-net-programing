package chapter09

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		/*
			could be replaced with
			// This function sets the content type to text/plain
			http.Error(w, "Bad request", http.StatusBadRequest)
		*/
		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
	}
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()
	handler(w, r)
	// Response status: "200 OK"
	// expected 400
	t.Logf("Response status: %q", w.Result().Status)

	handler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
	r = httptest.NewRequest(http.MethodGet, "http://test", nil)
	w = httptest.NewRecorder()
	handler(w, r)
	// Response status: "400 Bad Request"
	t.Logf("Response status: %q", w.Result().Status)
}
