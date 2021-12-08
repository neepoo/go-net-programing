package chapter13

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
)

func Test_Example_logMultiWriter(t *testing.T) {
	logFile := new(bytes.Buffer)
	w := SustainedMultiWriter(os.Stdout, logFile)
	l := log.New(w, "example: ", log.Lshortfile|log.Lmsgprefix)
	fmt.Println("standard output:")
	l.Print("Canada is south of Detroit")
	fmt.Print("\nlog file contents:\n", logFile.String())

}

func Test_Example_logLevels(t *testing.T) {
	lDebug := log.New(os.Stdout, "DEBUG: ", log.Lshortfile)
	logFile := new(bytes.Buffer)
	w := SustainedMultiWriter(logFile, lDebug.Writer())
	lError := log.New(w, "ERROR: ", log.Lshortfile)
	fmt.Println("standard output:")
	lError.Print("cannot communicate with the database")
	lDebug.Print("you cannot hum while holding your nose")
	fmt.Print("\nlog file contents:\n", logFile.String())
}
