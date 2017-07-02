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

//func NewWebsocketListener(laddr string) (net.Listener, error) {
//	return websocket.BinaryFrame
//}

func ListenUDP(laddr string) (net.Listener, error) {
	return net.Listen("udp", laddr)
}

func ListenTCP(laddr string) (net.Listener, error) {
	return net.Listen("tcp", laddr)
}

func ListenKCP(laddr string) (net.Listener, error) {
	return kcp.Listen(laddr)
}

func GetMessage(conn net.Conn) (*Message, error) {
	conn.SetDeadline(time.Now().Add(time.Second * 3))
	raw, err := read(conn)
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

func SendRaw(data []byte, conn net.Conn) error {
	conn.SetDeadline(time.Now().Add(time.Second * 3))
	size := len(data)
	if size > math.MaxUint16 {
		return fmt.Errorf("message size too large: %v", size)
	}
	sizeInBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeInBytes, uint16(size))
	_, err := conn.Write(append(sizeInBytes, data...))
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
