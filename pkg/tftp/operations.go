package tftp

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1
	_
	OpData
	OpAck
	OpErr
)

type Operation interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type Ack uint16

func (a Ack) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(OpCodeSize + BlockNumberSize)

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

func (e Err) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow((OpCodeSize + ErrCodeSize) + len(e.Message) + 1)

	err := binary.Write(b, binary.BigEndian, OpErr)
	if err != nil {
		return nil, fmt.Errorf("error writing operation code to buffer: [%w]", err)
	}

	err = binary.Write(b, binary.BigEndian, &e.Error)
	if err != nil {
		return nil, fmt.Errorf("error writing error code to buffer: [%w]", err)
	}

	_, err = io.CopyN(b, bytes.NewReader([]byte(e.Message)), DatagramSize)
	if err != nil {
		return nil, fmt.Errorf("error writing error message to buffer: [%w]", err)
	}

	return b.Bytes(), nil
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
	b.Grow((OpCodeSize + BlockNumberSize) + len(rrq.Filename) + 1 + len(rrq.Mode) + 1)

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

func UnmarshalBinary(buf []byte) (Operation, error) {
	var (
		ackPkt Ack
		errPkt Err
		rrq    ReadRequest
	)
	switch {
	case rrq.UnmarshalBinary(buf) == nil:
		return &rrq, nil
	case ackPkt.UnmarshalBinary(buf) == nil:
		return &ackPkt, nil
	case errPkt.UnmarshalBinary(buf) == nil:
		return &errPkt, nil
	default:
		return nil, fmt.Errorf("buffer could not be unmarshaled into operation")
	}
}
