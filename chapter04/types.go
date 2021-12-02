package chapter04

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize uint32 = 10 << 20 // 10MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom // read until eof
	io.WriterTo
	Bytes() []byte
}

type Binary []byte

func (b Binary) String() string {
	return string(b)
}

func (b *Binary) ReadFrom(r io.Reader) (n int64, err error) {
	var typ uint8
	err = binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}
	n = 1
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
	}
	var size uint32
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}
	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}
	*b = make([]byte, size)
	o, err := r.Read(*b)
	return n + int64(o), err
}

func (b Binary) WriteTo(w io.Writer) (n int64, err error) {
	// write header
	binary.Write(w, binary.BigEndian, BinaryType)
	if err != nil {
		return 0, err
	}
	n = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(b)))
	if err != nil {
		return 0, err
	}
	n += 4
	// write payload
	o, err := w.Write(b)
	return n + int64(o), err

}

func (b Binary) Bytes() []byte {
	return b
}

type String string

func (s String) String() string {
	return string(s)
}

func (s *String) ReadFrom(r io.Reader) (n int64, err error) {
	var typ uint8
	err = binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return
	}
	n = 1
	if typ != StringType {
		return n, errors.New("invalid String")
	}
	var size uint32
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}
	n += 4
	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}
	*s = String(buf)
	return n + int64(o), nil
}

func (s String) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.BigEndian, StringType)
	if err != nil {
		return
	}
	n = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(s)))
	if err != nil {
		return
	}
	n += 4
	o, err := w.Write(s.Bytes())
	return n + int64(o), err
}

func (s String) Bytes() []byte {
	return []byte(s)
}

func decode(r io.Reader) (Payload, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}
	_, err = payload.ReadFrom(io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}
	return payload, err
}
