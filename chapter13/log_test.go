package chapter13

import (
	"log"
	"os"
)

func Example_log() {
	l := log.New(os.Stdout, "example: ", log.Lshortfile)
	l.Print("Logging to standard output")
}
