package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"time"
)

type Server struct {
	Payload []byte        // the payload served for all read requests
	Retries uint8         // the number of times to retry a failed transmission
	Timeout time.Duration // the duration to wait for an acknowledgment
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func (s Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("Listening on %s ...\n", conn.LocalAddr())
	return s.Serve(conn)

}

func (s *Server) Serve(conn net.PacketConn) error {
	if conn == nil {
		return errors.New("nil connection")
	}
	if s.Payload == nil {
		return errors.New("payload is required")
	}
	if s.Retries == 0 {
		s.Retries = 10
	}
	if s.Timeout == 0 {
		s.Timeout = 6 * time.Second
	}

	var rrq ReadReq

	for {
		buf := make([]byte, DatagramSize)
		nr, addr, err := conn.ReadFrom(buf)
		log.Printf("local addr:[%s] %d bytes: [%q]", conn.LocalAddr().String(), nr, buf[:nr])
		if err != nil {
			// EOF or error
			log.Printf("rrq read error, err:%v\n", err)
			return err
		}
		err = rrq.UnmarshalBinary(buf)
		if err != nil {
			log.Printf("[%s] bad request: %v", addr, err)
			continue
		}
		log.Printf("first rrq: %#v\n", rrq)
		go s.handle(addr.String(), rrq)
	}
}

func (s Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)
	con, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial: %v", clientAddr, err)
		return
	}
	defer con.Close()

	var (
		ackPkt  Ack
		errPkt  Err
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)}
		buf     = make([]byte, DatagramSize)
	)
NEXTPACKET:
	for n := DatagramSize; n == DatagramSize; {
		data, err := dataPkt.MarshalBinary()
		if err != nil {
			log.Printf("[%s] preparing data packet: %v", clientAddr, err)
			return
		}
		log.Printf("data.block: %d\n", dataPkt.Block)
	RETRY:
		for i := s.Retries; i > 0; i-- {
			n, err = con.Write(data) // send the data packet
			if err != nil {
				log.Printf("[%s] write: %v", clientAddr, err)
				return
			}

			// wait for the client's ACK packet
			con.SetDeadline(time.Now().Add(s.Timeout))

			nr, err := con.Read(buf)
			if err != nil {
				if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
					continue RETRY
				}
				log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
				return
			}
			var oop uint16
			binary.Read(bytes.NewReader(buf[:2]), binary.BigEndian, &oop)
			log.Printf("receieved from client %d Bytes, op= %d, data: %q\n", nr, oop, buf[:nr])

			switch {
			case ackPkt.UnmarshalBinary(buf) == nil:
				if uint16(ackPkt) == dataPkt.Block {
					// received ACK; send next data packet
					continue NEXTPACKET
				}
			case errPkt.UnmarshalBinary(buf) == nil:
				log.Printf("[%s] received error: %v", clientAddr, errPkt.Message)
				return
			default:
				log.Printf("[%s] bad packet", clientAddr)
			}
		}
		log.Printf("[%s] exhausted retries", clientAddr)
		return
	}
	log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}
