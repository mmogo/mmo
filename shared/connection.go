package shared

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/xtaci/kcp-go"
)

const (
	ProtocolKCP = "kcp"
	ProtocolUDP = "udp"
	ProtocolTCP = "tcp"
)

func Dial(protocol, raddr string) (net.Conn, error) {
	switch protocol {
	case ProtocolKCP:
		return kcp.Dial(raddr)
	case ProtocolUDP:
		return net.Dial("udp", raddr)
	case ProtocolTCP:
		return net.Dial("tcp", raddr)
	}
	return nil, fmt.Errorf("invalid protcol %s. select from available: %s | %s | %s", ProtocolKCP, ProtocolUDP, ProtocolTCP)
}

func Listen(protocol, laddr string) (net.Listener, error) {
	switch protocol {
	case ProtocolKCP:
		return kcp.Listen(laddr)
	case ProtocolUDP:
		return net.Listen("udp", laddr)
		//case ProtocolTCP:
		//	return net.Listen("tcp", laddr)
	}
	return nil, fmt.Errorf("invalid protcol %s. select from available: %s | %s | %s", ProtocolKCP, ProtocolUDP, ProtocolTCP)
}

func GetMessage(r io.Reader) (*Message, error) {
	raw, err := read(r)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(raw))
	var msg Message
	if err := dec.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func SendMessage(msg *Message, conn net.Conn) error {
	msg.Sent = time.Now()
	data, err := Encode(msg)
	if err != nil {
		return err
	}
	return SendRaw(data, conn)
}

func SendRaw(data []byte, w io.Writer) error {
	size := len(data)
	if size > math.MaxUint16 {
		return fmt.Errorf("message size too large: %v", size)
	}
	sizeInBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeInBytes, uint16(size))
	_, err := w.Write(append(sizeInBytes, data...))
	return err
}

func Encode(e interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(e)
	return buf.Bytes(), err
}

func read(r io.Reader) ([]byte, error) {
	sizeInBytes := make([]byte, 2)
	if _, err := r.Read(sizeInBytes); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint16(sizeInBytes)
	data := make([]byte, size)
	_, err := r.Read(data)
	return data, err
}
