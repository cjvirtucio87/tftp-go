package tftp

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	BlockNumberSize = 2
	DatagramSize    = 516
	ErrCodeSize     = 2
	BlockSize       = DatagramSize - (OpCodeSize + BlockNumberSize)
	OpCodeSize      = 2
)

type Server struct {
	Logger  Logger
	Payload []byte
	Retries uint8
	Timeout time.Duration
}

func (s Server) handle(conn net.PacketConn, addr net.Addr, op Operation) {
	switch op.(type) {
	case *ReadRequest:
		dataPkt := Data{
			Payload: bytes.NewReader(s.Payload),
		}
		err := sendDataPkt(conn, addr, dataPkt)
		if err != nil {
			log.Printf("error sending data packet to client [%s]: %v", addr.String(), err)
		}

		return
	case *Ack:
		dataPkt := Data{
			Payload: bytes.NewReader(s.Payload),
		}

		ackPkt, ok := op.(*Ack)
		if !ok {
			return
		}

		if uint16(*ackPkt) != dataPkt.Block {
			return
		}

		err := sendDataPkt(conn, addr, dataPkt)
		if err != nil {
			log.Printf("error sending data packet to client [%s]: %v", addr.String(), err)
			return
		}
	case *Err:
		errPkt, ok := op.(*Err)
		if !ok {
			return
		}

		s.Logger.Errorf("[%s] received error: %v", addr.String(), errPkt.Message)
		return
	default:
		s.Logger.Infof("[%s] bad packet", addr.String())
	}
}

func (s Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: [%w]", err)
	}

	defer func(conn net.PacketConn) {
		_ = conn.Close()
	}(conn)

	s.Logger.Infof("Listening on %s ...", conn.LocalAddr())

	return s.Serve(conn)
}

func sendDataPkt(conn net.PacketConn, addr net.Addr, dataPkt Data) error {
	data, err := dataPkt.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error during attempt to send data packet: %w", err)
	}

	_, err = conn.WriteTo(data, addr)
	if err != nil {
		return fmt.Errorf("error during attempt to send data packet: %w", err)
	}

	return nil
}

func (s Server) Serve(conn net.PacketConn) error {
	if conn == nil {
		return fmt.Errorf("conn must not be nil")
	}

	if s.Payload == nil {
		return fmt.Errorf("payload is required")
	}

	if s.Retries == 0 {
		s.Retries = 10
	}

	if s.Timeout == 0 {
		s.Timeout = 6 * time.Second
	}

	for {
		buf := make([]byte, DatagramSize)

		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return fmt.Errorf("failed to read request into buffer: [%w]", err)
		}

		op, err := UnmarshalBinary(buf)
		if err != nil {
			return fmt.Errorf("error unmarshaling binary from client [%s]: [%w]", conn.LocalAddr(), err)
		}

		go s.handle(conn, addr, op)
	}
}
