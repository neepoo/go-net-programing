package chapter04

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"reflect"
	"testing"
)

func TestPayload(t *testing.T) {
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")
	s1 := String("Errors are values.")
	payloads := []Payload{&b1, &s1, &b2}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		for _, p := range payloads {
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}

	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < len(payloads); i++ {
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}

		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}
		t.Logf("[%T] %[1]q", actual)
	}
}

func TestMaxPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer)
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1GB
	if err != nil {
		t.Fatal(err)
	}
	var b Binary
	_, err = b.ReadFrom(buf)
	if !errors.Is(err, ErrMaxPayloadSize) {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
