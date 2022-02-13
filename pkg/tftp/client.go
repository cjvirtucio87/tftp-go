package tftp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

type Client struct {
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

func (c Client) sendRrq(conn net.PacketConn, addr string, filename string) error {
	b, err := ReadRequest{
		Filename: filename,
		Mode:     "octet",
	}.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to create read request: [%w]", err)
	}

	err = c.send(
		conn,
		addr,
		b,
	)
	if err != nil {
		return fmt.Errorf("failed sending read request: [%w]", err)
	}

	for i := 10; i > 0; i-- {
		replyBuf := make([]byte, DatagramSize)
		_, _, err = conn.ReadFrom(replyBuf)
		if err != nil {
			return fmt.Errorf("[%s] error reading reply from server: [%w]", conn.LocalAddr(), err)
		}

		var dataPkt Data
		err = dataPkt.UnmarshalBinary(replyBuf)
		if err != nil {
			log.Printf("[%s] error unmarshaling data packet from server: [%v]", conn.LocalAddr(), err)
			continue
		}

		payloadBuf := new(bytes.Buffer)
		_, err = io.CopyN(payloadBuf, dataPkt.Payload, BlockSize)
		if err != nil {
			log.Printf("[%s] error reading payload into buffer: [%v]", conn.LocalAddr(), err)
			continue
		}

		log.Printf("[%s] reply from addr [%s]: [%s]", conn.LocalAddr(), addr, string(payloadBuf.Bytes()))
	}

	return nil
}
