package chapter13

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
)

// Wide event logging is a technique that creates a single, structured log entry
//per event to summarize the transaction, instead of logging numerous entries
//as the transaction progresses. This technique is most applicable to request-
//response loops, such as API calls, but it can be adapted to other use cases.
//One approach to wide event logging is to wrap an API handler in middleware.

type wideResponseWriter struct {
	http.ResponseWriter
	length int // content-length
	status int // status code
}

func (w *wideResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func (w *wideResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return n, err
}

func WideEventLog(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			wideWriter := &wideResponseWriter{ResponseWriter: w}
			next.ServeHTTP(wideWriter, r)
			addr, _, _ := net.SplitHostPort(r.RemoteAddr)
			// 日志会在响应放回之前输出
			logger.Info(
				"example wide event",
				zap.Int("status_code", wideWriter.status),
				zap.Int("response_length", wideWriter.length),
				zap.Int64("request_content_length", r.ContentLength),
				zap.String("method", r.Method),
				zap.String("proto", r.Proto),
				zap.String("remote_addr", addr),
				zap.String("uri", r.RequestURI),
				zap.String("user_agent", r.UserAgent()))
		},
	)
}

func Example_wideLogEntry() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		),
	)
	defer zl.Sync()

	ts := httptest.NewServer(
		WideEventLog(zl,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func(r io.ReadCloser) {
					io.Copy(ioutil.Discard, r)
					r.Close()
				}(r.Body)
				w.Write([]byte("hello!"))
			},
			),
		),
	)

	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if err != nil {
		zl.Fatal(err.Error())
	}
	resp.Body.Close()

	// Output:
	// {"level":"info","msg":"example wide event","status_code":200,"response_length":6,"request_content_length":0,"method":"GET","proto":"HTTP/1.1","remote_addr":"127.0.0.1","uri":"/","user_agent":"Go-http-client/1.1"}
}
