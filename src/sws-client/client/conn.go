package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	reader io.Reader
	wsConn *websocket.Conn
}

func NewConn(wsConn *websocket.Conn) *Conn {
	return &Conn{
		wsConn: wsConn,
	}
}

// Read implements the net.Conn.Read()
func (c *Conn) Read(b []byte) (int, error) {
	var N int
	capacity := cap(b)
	for {
		reader, err := c.getReader()
		if err != nil {
			return 0, err
		}

		n, err := reader.Read(b[N:])
		N += n
		if err == io.EOF {
			c.reader = nil
			if N < capacity {
				continue
			}
		}
		return N, err
	}
}

func (c *Conn) getReader() (io.Reader, error) {
	if c.reader != nil {
		return c.reader, nil
	}

	_, reader, err := c.wsConn.NextReader()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	l := len(b)
	r := make([]byte, l-1)
	for i := 1; i < l; i++ {
		r[l-1-i] = b[i] - b[0]
	}
	c.reader = bytes.NewReader(r)
	return c.reader, nil
}

// Write implements the io.Writer.
func (c *Conn) Write(b []byte) (int, error) {
	l := len(b)
	if l == 0 {
		return 0, nil
	}

	r := make([]byte, l+1)
	rc := make([]byte, 1)
	rand.Read(rc)
	r[0] = rc[0]
	for i := 0; i < l; i++ {
		r[l-i] = b[i] + rc[0]
	}

	err := c.wsConn.WriteMessage(websocket.BinaryMessage, r)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func (c *Conn) Close() error {
	c.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(3*time.Second))
	return c.wsConn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.wsConn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.wsConn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.wsConn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.wsConn.SetWriteDeadline(t)
}
