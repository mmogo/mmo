package shared

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/xtaci/kcp-go"
	"gopkg.in/mgo.v2/bson"
)

const (
	ProtocolUDP = "udp"
	ProtocolTCP = "tcp"
)

func Dial(protocol, raddr string) (net.Conn, error) {
	switch protocol {
	case ProtocolUDP:
		return kcp.Dial(raddr)
	case ProtocolTCP:
		return net.Dial("tcp", raddr)
	}
	return nil, fmt.Errorf("invalid protcol %s. select from available: %s | %s", ProtocolUDP, ProtocolTCP)
}

func Listen(protocol, laddr string) (net.Listener, error) {
	switch protocol {
	case ProtocolUDP:
		return kcp.Listen(laddr)
	case ProtocolTCP:
		return net.Listen("tcp", laddr)
	}
	return nil, fmt.Errorf("invalid protcol %s. select from available: %s | %s ", ProtocolUDP, ProtocolTCP)
}

func GetMessage(r io.Reader, withDeadline ...bool) (*Message, error) {
	if len(withDeadline) > 0 && withDeadline[0] {
		if conn, ok := r.(net.Conn); ok {
			conn.SetDeadline(time.Now().Add(time.Second * 3))
		}
	}
	raw, err := read(r)
	if err != nil {
		return nil, err
	}
	var msg Message
	if err := bson.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func SendMessage(msg *Message, w io.Writer, withDeadline ...bool) error {
	if len(withDeadline) > 0 && withDeadline[0] {
		if conn, ok := w.(net.Conn); ok {
			conn.SetDeadline(time.Now().Add(time.Second * 3))
		}
	}
	msg.Sent = time.Now()
	data, err := Encode(msg)
	if err != nil {
		return err
	}
	return SendRaw(data, w)
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
	return bson.Marshal(e)
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
