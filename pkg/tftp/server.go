package tftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	DatagramSize = 516
	BlockSize    = DatagramSize - 4 // accounting for OpCode and Block number
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1
	_            // write request unsupported for this exercise
	OpData
	OpAck
	OpErr
)

type Ack uint16

// first 2 bytes: op code
// last 2 bytes: block number for the data block that the client is acknowledge receipt of
func (a Ack) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(2 + 2)

	err := binary.Write(b, binary.BigEndian, OpAck)
	if err != nil {
		return nil, fmt.Errorf("error writing acknowledgement operation code to binary: [%w]", err)
	}

	err = binary.Write(b, binary.BigEndian, a)
	if err != nil {
		return nil, fmt.Errorf("error acknowledgement operation code to binary: [%w]", err)
	}

	return b.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(b []byte) error {
	var code OpCode
	r := bytes.NewReader(b)

	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return fmt.Errorf("encountered error reading binary into operation code: [%w]", err)
	}

	if code != OpAck {
		return fmt.Errorf("invalid code for acknowledgement packet: [%d]", code)
	}

	return binary.Read(r, binary.BigEndian, a)
}

type Data struct {
	Block   uint16
	Payload io.Reader
}

// 2 bytes OpCode
// 2 bytes Block
// n bytes Payload
func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++

	err := binary.Write(b, binary.BigEndian, OpData)
	if err != nil {
		return nil, fmt.Errorf("error writing operation code: [%w]", err)
	}

	err = binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, fmt.Errorf("error writing block number: [%w]", err)
	}

	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error writing payload up to block size: [%w]", err)
	}

	return b.Bytes(), nil
}

func (d *Data) UnmarshalBinary(b []byte) error {
	l := len(b)

	if l < 4 {
		return fmt.Errorf("missing header bytes in binary")
	}

	if l > DatagramSize {
		return fmt.Errorf("binary size [%d] exceeds DatagramSize limit", l)
	}

	var code OpCode
	err := binary.Read(bytes.NewReader(b[:2]), binary.BigEndian, &code)
	if err != nil {
		return fmt.Errorf("encountered error reading binary into operation code: [%w]", err)
	}

	if code != OpData {
		return fmt.Errorf("expected data code, got [%d]", code)
	}

	err = binary.Read(bytes.NewReader(b[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return fmt.Errorf("encountered error reading binary into block number: [%w]", err)
	}

	d.Payload = bytes.NewReader(b[4:])

	return nil
}

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

type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) UnmarshalBinary(b []byte) error {
	var code OpCode
	r := bytes.NewBuffer(b)

	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return fmt.Errorf("encountered error reading binary into operation code: [%w]", err)
	}

	if code != OpErr {
		return fmt.Errorf("invalid code for error packet: [%d]", code)
	}

	err = binary.Read(r, binary.BigEndian, &e.Error)
	if err != nil {
		return fmt.Errorf("error attempting to unmarshal binary into ErrCode: [%w]", err)
	}

	e.Message, err = r.ReadString(0)
	e.Message = strings.TrimRight(e.Message, "\x00")

	return err
}

type ReadRequest struct {
	Filename string
	Mode     string
}

func (rrq ReadRequest) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(2 + 2 + len(rrq.Filename) + 1 + len(rrq.Mode) + 1)

	err := binary.Write(b, binary.BigEndian, OpRRQ)
	if err != nil {
		return nil, fmt.Errorf("failed to write operation code to bytes buffer: [%w]", err)
	}

	_, err = b.WriteString(rrq.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to write filename to bytes buffer: [%w]", err)
	}

	err = b.WriteByte(0)
	if err != nil {
		return nil, fmt.Errorf("failed to zero-byte delimiter for read request binary: [%w]", err)
	}

	_, err = b.WriteString(rrq.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to write mode to bytes buffer: [%w]", err)
	}

	err = b.WriteByte(0)
	if err != nil {
		return nil, fmt.Errorf("failed to zero-byte delimiter for read request binary: [%w]", err)
	}

	return b.Bytes(), nil
}

// first 2 bytes: operation code
// next n bytes: filename
// 0 byte delimiter
// next n bytes: mode
// 0 byte delimiter
func (rrq *ReadRequest) UnmarshalBinary(b []byte) error {
	r := bytes.NewBuffer(b)

	var code OpCode
	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return fmt.Errorf("binary does not contain OpCode header: [%w]", err)
	}

	if code != OpRRQ {
		return fmt.Errorf("invalid code for read request packet: [%d]", code)
	}

	rrq.Filename, err = r.ReadString(0)
	if err != nil {
		return fmt.Errorf("error reading filename: [%w]", err)
	}

	rrq.Filename = strings.TrimRight(rrq.Filename, "\x00")
	if len(rrq.Filename) == 0 {
		return fmt.Errorf("invalid filename: [%s]", rrq.Filename)
	}

	rrq.Mode, err = r.ReadString(0)
	if err != nil {
		return fmt.Errorf("invalid mode: [%s]", rrq.Mode)
	}

	rrq.Mode = strings.TrimRight(rrq.Mode, "\x00")
	if len(rrq.Mode) == 0 {
		return fmt.Errorf("invalid mode: [%s]", rrq.Mode)
	}

	if !strings.EqualFold("octet", rrq.Mode) {
		return fmt.Errorf("unsupported read request mode: [%s]", rrq.Mode)
	}

	return nil
}

type Server struct {
    Logger Logger
	Payload []byte
	Retries uint8
	Timeout time.Duration
}

func (s Server) handle(conn net.PacketConn, addr net.Addr, buf []byte) {
	var (
		ackPkt  Ack
		dataPkt Data
		errPkt  Err
		rrq     ReadRequest
	)
	switch {
	case rrq.UnmarshalBinary(buf) == nil:
		dataPkt = Data{
			Payload: bytes.NewReader(s.Payload),
		}
		err := sendDataPkt(conn, addr, dataPkt)
		if err != nil {
			log.Printf("error sending data packet to client [%s]: %v", addr.String(), err)
		}

		return
	case ackPkt.UnmarshalBinary(buf) == nil:
		dataPkt = Data{
			Payload: bytes.NewReader(s.Payload),
		}
		if uint16(ackPkt) != dataPkt.Block {
			return
		}

		err := sendDataPkt(conn, addr, dataPkt)
		if err != nil {
			log.Printf("error sending data packet to client [%s]: %v", addr.String(), err)
			return
		}
	case errPkt.UnmarshalBinary(buf) == nil:
		log.Printf("[%s] received error: %v", addr.String(), errPkt.Message)
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

		go s.handle(conn, addr, buf)
	}
}
