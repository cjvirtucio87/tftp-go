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
  BlockSize = DatagramSize - 4 // accounting for OpCode and Block number
)

type OpCode uint16

const (
  OpRRQ OpCode = iota + 1
  _ // write request unsupported for this exercise
  OpData
  OpAck
  OpErr
)

type Ack uint16

// first 2 bytes: op code
// last 2 bytes: block number for the data block that the client is acknowledge receipt of
func (a Ack) MarshalBinary() ([]byte, error) {
  b := new(bytes.Buffer)
  b.Grow(DatagramSize)

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

func (a Ack) UnmarshalBinary(b []byte) error {
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
  Block uint16
  Payload io.Reader
}

// 2 bytes OpCode
// 2 bytes Block
// n bytes Payload
func (d Data) MarshalBinary() ([]byte, error) {
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
  if err != nil {
    return nil, fmt.Errorf("error writing payload up to block size: [%w]", err)
  }

  return b.Bytes(), nil
}

func (d Data) UnmarshalBinary(b []byte) error {
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
  Error ErrCode
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
  Mode string
}

// first 2 bytes: operation code
// next n bytes: filename
// 0 byte delimiter
// next n bytes: mode
// 0 byte delimiter
func (rrq ReadRequest) UnmarshalBinary(b []byte) error {
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

  if strings.ToLower(rrq.Mode) != "octet" {
    return fmt.Errorf("only binary transfers supported")
  }

  return nil
}

type Server struct {
  Payload []byte
  Retries uint8
  Timeout time.Duration
}

func (s Server) handle(clientAddr string, rrq ReadRequest) {
  log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)

  conn, err := net.Dial("udp", clientAddr)
  if err != nil {
    log.Printf("[%s] error dialing client address: %v", clientAddr, err)
    return
  }

  defer func() {
    _ = conn.Close()
  }()

  var (
    ackPkt Ack
    errPkt Err
    dataPkt = Data{
      Payload: bytes.NewReader(s.Payload),
    }
    buf = make([]byte, DatagramSize)
  )

  NEXTPACKET:
  for n := DatagramSize; n == DatagramSize; {
      data, err := dataPkt.MarshalBinary()
      if err != nil {
          log.Printf("[%s] preparing data packet: %v", clientAddr, err)
          return
      }

      RETRY:
      for i := s.Retries; i > 0; i-- {
          n, err = conn.Write(data) // send the data packet
          if err != nil {
              log.Printf("[%s] write: %v", clientAddr, err)
              return
          }

          // wait for the client's ACK packet
          _ = conn.SetReadDeadline(time.Now().Add(s.Timeout))

          _, err = conn.Read(buf)
          if err != nil {
              if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
                  continue RETRY
              }

              log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
              return
          }

          switch {
          case ackPkt.UnmarshalBinary(buf) == nil:
              if uint16(ackPkt) == dataPkt.Block {
                  // received ACK; send next data packet
                  continue NEXTPACKET
              }
          case errPkt.UnmarshalBinary(buf) == nil:
              log.Printf("[%s] received error: %v",
                  clientAddr, errPkt.Message)
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

func (s Server) ListenAndServe(addr string) error {
  conn, err := net.ListenPacket("udp", addr)
  if err != nil {
    return fmt.Errorf("failed to start server: [%w]", err)
  }

  defer func(conn net.PacketConn) {
    _ = conn.Close()
  }(conn)

  log.Printf("Listening on %s ...\n", conn.LocalAddr())

  return s.Serve(conn)
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

    var rrq ReadRequest
    if err = rrq.UnmarshalBinary(buf); err != nil {
      log.Printf("[%s] bad address: [%v]", addr, err)
      continue
    }

    go s.handle(addr.String(), rrq)
  }
}
