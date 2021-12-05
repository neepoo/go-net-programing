package handlers

import (
	"database/sql"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

type Methods map[string]http.Handler

func (h Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(r io.ReadCloser) {
		io.Copy(io.Discard, r)
		r.Close()
	}(r.Body)

	if handler, ok := h[r.Method]; ok {
		if handler == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} else {
			handler.ServeHTTP(w, r)
		}
		return
	}
	w.Header().Add("Allow", h.allowedMethods())
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allow", http.StatusMethodNotAllowed)
	}
}

func (h Methods) allowedMethods() string {
	a := make([]string, 0, len(h))
	for k := range h {
		a = append(a, k)
	}
	sort.Strings(a)
	return strings.Join(a, ", ")
}

func DefaultMethodHandler() http.Handler {
	return Methods{
		http.MethodGet: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello friend!"))
		}),
		http.MethodPost: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "Hello, %s!", html.EscapeString(string(b)))
		}),
	}
}

// how to inject a SQL database object into an http.Handle

var dbHandler func(db *sql.DB) http.Handler = func(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			return
		}
	})
}

// Handlers example http dependency inject
type Handlers struct {
	db  *sql.DB
	log *log.Logger
}

func (h *Handlers) Handler1() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := h.db.Ping()
			if err != nil {
				h.log.Printf("db ping: %v", err)
			}
			// do something with the database here
		},
	)
}
func (h *Handlers) Handler2() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// ...
		},
	)
}

//var h = &Handlers{
//	db:  db,
//	log: log.New(os.Stderr, "handlers: ", log.Lshortfile),
//}
//http.Handle("/one", h.Handler1())
//http.Handle("/two", h.Handler2())
