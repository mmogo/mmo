package shared

import (
	"bytes"
	"encoding/gob"
	"github.com/gorilla/websocket"
)

func GetMessage(conn *websocket.Conn) (*Message, error) {
	_, raw, err := conn.ReadMessage()
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

func SendMessage(msg *Message, conn *websocket.Conn) error {
	data, err := Encode(msg)
	if err != nil {
		return err
	}
	return SendRaw(data, conn)
}

func SendRaw(data []byte, conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.BinaryMessage, data)
}

func Encode(e interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(e)
	return buf.Bytes(), err
}
