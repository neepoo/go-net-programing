package main

import (
	"flag"
	"go-network/chapter06/tftp"
	"io/ioutil"
	"log"
)

var (
	address = flag.String("a", "127.0.0.1:69", "listen address")
	payload = flag.String("p", "payload.svg", "file to serve to client")
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	flag.Parse()

	p, err := ioutil.ReadFile(*payload)
	if err != nil {
		log.Fatal(err)
	}
	s := tftp.Server{Payload: p}
	log.Fatal(s.ListenAndServe(*address))
}
