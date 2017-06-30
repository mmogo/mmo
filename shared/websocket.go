package shared

import (
	"net"

	"context"
	"net/http"
	"time"

	"fmt"

	"github.com/gorilla/websocket"
)

type wsConn struct {
	*websocket.Conn
}

func (c *wsConn) Write(p []byte) (n int, err error) {
	return len(p), c.WriteMessage(websocket.BinaryMessage, p)
}

func (c *wsConn) Read(p []byte) (n int, err error) {
	_, data, err := c.ReadMessage()
	if err != nil {
		return len(data), err
	}
	size := 0
	for i := range p {
		if i == len(data) {
			break
		}
		p[i] = data[i]
		size++
	}
	return size, nil
}

func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

type result struct {
	conn *wsConn
	err  error
}

type websocketListener struct {
	server  *http.Server
	results chan result
	addr    net.Addr
}

func ListenWS(laddr string) (net.Listener, error) {
	mux := http.NewServeMux()
	results := make(chan result)
	l := &websocketListener{}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, req, nil)
		//set addr easy the first time
		if l.addr == nil {
			l.addr = conn.LocalAddr()
		}
		if err != nil {
			results <- result{
				conn: nil,
				err:  fmt.Errorf("failed to upgrade connection for %v: %v", req, err),
			}
			return
		}
		results <- result{
			conn: &wsConn{conn},
			err:  nil,
		}
	})
	server := &http.Server{
		Addr:    laddr,
		Handler: mux,
	}
	go func() {
		server.ListenAndServe()
	}()
	l.server = server
	l.results = results
	return l, nil
}

func (l *websocketListener) Accept() (net.Conn, error) {
	res := <-l.results
	return res.conn, res.err
}

func (l *websocketListener) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return l.server.Shutdown(ctx)
}

func (l *websocketListener) Addr() net.Addr {
	return l.addr
}
