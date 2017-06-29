package shared

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

func GetMessage(conn net.Conn) (*Message, error) {
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
	data, err := Encode(msg)
	if err != nil {
		return err
	}
	return SendRaw(data, conn)
}

func SendRaw(data []byte, conn net.Conn) error {
	size := uint16(len(data))
	sizeInBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeInBytes, size)
	buf := append(sizeInBytes, data...)
	_, err := conn.Write(buf)
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
