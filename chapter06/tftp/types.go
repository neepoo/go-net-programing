package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

// Your TFTP server will accept read requests from the client, send data pack-
// ets, transmit error packets, and accept acknowledgments from the client

const (
	DatagramSize = 516              // the maximum supported datagram size. avoid fragmentation
	BlockSize    = DatagramSize - 4 // the DatagramSize minus a 4-byte header
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1 // Read Request
	OpWRQ
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

const (
	ErrUnknown ErrCode = iota
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownID
	ErrFileExists
	ErrNoUser
)

// ReadReq 客户端读(下载)文件请求
type ReadReq struct {
	Filename string
	Mode     string // only support octet
}

func (q ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != mode {
		mode = q.Mode
	}
	// opCode      filename          \0   mode         \0
	c := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1
	b := new(bytes.Buffer)
	b.Grow(c)

	err := binary.Write(b, binary.BigEndian, OpRRQ) // write op code
	if err != nil {
		return nil, err
	}
	_, err = b.WriteString(q.Filename) // write filename
	if err != nil {
		return nil, err
	}
	err = b.WriteByte(0) // write 0 byte
	if err != nil {
		return nil, err
	}

	// write mode
	_, err = b.WriteString(q.Mode)
	if err != nil {
		return nil, err
	}

	err = b.WriteByte(0) // write 0
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (q *ReadReq) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)
	var code OpCode

	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}
	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}
	q.Filename = strings.TrimRight(q.Filename, "\x00") // remove the 0-byte
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}
	q.Mode, err = r.ReadString(0) // read mode
	if err != nil {
		return errors.New("invalid RRQ")
	}
	q.Mode = strings.TrimRight(q.Mode, "\x00") // remove the 0-byte
	if len(q.Mode) == 0 {
		return errors.New("invalid RRQ")
	}
	actual := strings.ToLower(q.Mode) // enforce octet mode
	if actual != "octet" {
		return errors.New("only binary transfers supported")
	}

	return nil
}

type Data struct {
	Block   uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++
	err := binary.Write(b, binary.BigEndian, OpData)
	if err != nil {
		return nil, err
	}
	err = binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return b.Bytes(), nil
}

func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}
	var opcode OpCode

	//reader := bytes.NewReader(p)
	//err := binary.Read(reader, binary.BigEndian, &opcode)
	//if err != nil || opcode != OpData {
	//	return errors.New("invalid DATA")
	//}
	//
	//err = binary.Read(reader, binary.BigEndian, &d.Block)
	//if err != nil {
	//	return errors.New("invalid DATA")
	//}
	//d.Payload = reader

	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
	if err != nil || opcode != OpData {
		return errors.New("invalid DATA")
	}
	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}
	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

type Ack uint16

func (a *Ack) UnmarshalBinary(p []byte) error {
	var code OpCode

	r := bytes.NewReader(p)
	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpAck {
		return errors.New("invalid ACK")
	}
	return binary.Read(r, binary.BigEndian, a)
}

func (a Ack) MarshalBinary() ([]byte, error) {
	c := 2 + 2 // opcode + block num
	b := new(bytes.Buffer)
	b.Grow(c)

	err := binary.Write(b, binary.BigEndian, OpAck)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, a)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), err
}

type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	// operation code + error code + message + 0 byte
	c := 2 + 2 + len(e.Message) + 1
	b := new(bytes.Buffer)
	b.Grow(c)
	err := binary.Write(b, binary.BigEndian, OpErr) // write operation code
	if err != nil {
		return nil, err
	}
	err = binary.Write(b, binary.BigEndian, e.Error) // write error code
	if err != nil {
		return nil, err
	}
	_, err = b.WriteString(e.Message) // write message
	if err != nil {
		return nil, err
	}
	err = b.WriteByte(0) // write 0 byte
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *Err) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)
	var code OpCode
	err := binary.Read(r, binary.BigEndian, &code) // read operation code
	if err != nil {
		return err
	}
	if code != OpErr {
		return errors.New("invalid ERROR")
	}
	err = binary.Read(r, binary.BigEndian, &e.Error) // read error message
	if err != nil {
		return err
	}
	e.Message, err = r.ReadString(0)
	e.Message = strings.TrimRight(e.Message, "\x00") // remove the 0-byte
	return err
}
