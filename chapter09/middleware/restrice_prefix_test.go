package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestrictPrefix(t *testing.T) {
	handler := http.StripPrefix("/static/", RestrictPrefix(".", http.FileServer(http.Dir("../files/"))))
	testCase := []struct {
		path string
		code int
	}{
		{"http://test22/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	for i, c := range testCase {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		actual := w.Result().StatusCode
		if c.code != actual {
			t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
		}
	}
}
