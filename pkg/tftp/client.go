package tftp

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Client struct {
	Retries int
	Writer  io.Writer
}

func (c Client) Send(clientAddr string, addr string, filename string) error {
	conn, err := net.ListenPacket("udp", clientAddr)
	if err != nil {
		return fmt.Errorf("unable to listen on UDP address: [%s]", clientAddr)
	}

	err = c.sendRrq(conn, addr, filename)
	if err != nil {
		return fmt.Errorf("failed to send read request: [%w]", err)
	}

	n := int64(DatagramSize)
	for n == DatagramSize {
		dataPkt, err := c.waitForData(conn)
		if err != nil {
			return fmt.Errorf("[%s] error waiting for data packet: [%w]", conn.LocalAddr(), err)
		}

		n, err = io.Copy(c.Writer, dataPkt.Payload)
		if err != nil {
			return fmt.Errorf("[%s] error reading payload into writer: [%w]", conn.LocalAddr(), err)
		}

		err = c.sendAck(conn, addr, *dataPkt)
		if err != nil {
			return fmt.Errorf("[%s] error sending acknowledgement: [%w]", conn.LocalAddr(), err)
		}
	}

	return nil
}

func (c Client) send(conn net.PacketConn, addr string, b []byte) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("cannot resolve udp addr [%s]", addr)
	}

	n, err := conn.WriteTo(b, udpAddr)
	if err != nil {
		return fmt.Errorf("failed sending bytes to addr [%s]", udpAddr)
	}

	log.Printf("wrote [%d] bytes to addr [%s]", n, addr)

	return nil
}

func (c Client) sendAck(conn net.PacketConn, addr string, dataPkt Data) error {
	var ackPkt Ack
	ackPkt = Ack(dataPkt.Block)
	b, err := ackPkt.MarshalBinary()

	if err != nil {
		return fmt.Errorf("failed to send acknowledgement: [%w]", err)
	}

	return c.send(
		conn,
		addr,
		b,
	)
}

func (c Client) sendRrq(conn net.PacketConn, addr string, filename string) error {
	b, err := ReadRequest{
		Filename: filename,
		Mode:     "octet",
	}.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to create read request: [%w]", err)
	}

	return c.send(
		conn,
		addr,
		b,
	)
}

func (c Client) waitForData(conn net.PacketConn) (*Data, error) {
	for i := c.Retries; i > 0; i-- {
		replyBuf := make([]byte, DatagramSize)
		_, _, err := conn.ReadFrom(replyBuf)
		if err != nil {
			log.Printf("[%s] error reading reply from server: [%v]", conn.LocalAddr(), err)
			continue
		}

		var dataPkt Data
		err = dataPkt.UnmarshalBinary(replyBuf)
		if err != nil {
			log.Printf("[%s] error unmarshaling data packet from server: [%v]", conn.LocalAddr(), err)
			continue
		}

		return &dataPkt, nil
	}

	return nil, fmt.Errorf("[%s] ran out of retries waiting for data", conn.LocalAddr())
}
