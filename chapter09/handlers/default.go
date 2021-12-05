package handlers

import (
	"io"
	"io/ioutil"
	"net/http"
	"text/template"
)

var t = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

func DefaultHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func(r io.ReadCloser) {
				io.Copy(ioutil.Discard, r)
				r.Close()
			}(r.Body)

			var b []byte
			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")
			case http.MethodPost:
				var err error
				b, err = ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			default:
				http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
				return
			}
			_ = t.Execute(w, string(b))

		},
	)
}
