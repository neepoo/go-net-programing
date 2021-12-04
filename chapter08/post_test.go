package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type User struct {
	First string
	Last  string
}

func handlePostUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(r io.ReadCloser) {
			io.Copy(ioutil.Discard, r)
			r.Close()
		}(r.Body)

		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			t.Error(err)
			http.Error(w, "Decode Failed", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func TestPostUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handlePostUser(t)))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusMethodNotAllowed, resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	u := User{
		First: "zhikai",
		Last:  "wei",
	}
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusAccepted, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestMultipartPost(t *testing.T) {
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)
	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}
	// 添加文件
	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
		//"./files/v3.png",
	} {
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i+1), filepath.Base(file))
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(filePart, f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	/*
		--2088ecdb2bda83b5b8fc885bdeb4eaf1ea171134d03889b3f33447fd46d8
		Content-Disposition: form-data; name="date"

		2021-12-05T00:29:53+08:00
		--2088ecdb2bda83b5b8fc885bdeb4eaf1ea171134d03889b3f33447fd46d8
		Content-Disposition: form-data; name="description"

		Form values with attached files
		--2088ecdb2bda83b5b8fc885bdeb4eaf1ea171134d03889b3f33447fd46d8
		Content-Disposition: form-data; name="file1"; filename="hello.txt"
		Content-Type: application/octet-stream

		hello
		--2088ecdb2bda83b5b8fc885bdeb4eaf1ea171134d03889b3f33447fd46d8
		Content-Disposition: form-data; name="file2"; filename="goodbye.txt"
		Content-Type: application/octet-stream

		goodbye
	*/
	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://httpbin.org/post", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
	/*
	   {
	      "args": {},
	      "data": "",
	      "files": {
	        "file1": "hello",
	        "file2": "goodbye"
	      },
	      "form": {
	        "date": "2021-12-05T00:29:53+08:00",
	        "description": "Form values with attached files"
	      },
	      "headers": {
	        "Accept-Encoding": "gzip",
	        "Content-Length": "721",
	        "Content-Type": "multipart/form-data; boundary=2088ecdb2bda83b5b8fc885bdeb4eaf1ea171134d03889b3f33447fd46d8",
	        "Host": "httpbin.org",
	        "User-Agent": "Go-http-client/2.0",
	        "X-Amzn-Trace-Id": "Root=1-61ab97a4-7aa4856a34c6621f2536181f"
	      },
	      "json": null,
	      "origin": "122.118.121.4",
	      "url": "https://httpbin.org/post"
	    }

	*/
	t.Logf("\n%s", b)
}
