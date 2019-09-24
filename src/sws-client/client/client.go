package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	WSUri string
	Port  uint16
}

func (c *Client) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return
	}

	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			return
		}

		err = conn.(*net.TCPConn).SetKeepAlive(true)
		if err != nil {
			log.Println("[ERR] shadowX: failed to set KeepAlive on TCP connection")
			return
		}

		err = conn.(*net.TCPConn).SetKeepAlivePeriod(time.Hour)
		if err != nil {
			log.Println("[ERR] shadowX: failed to set KeepAlivePeriod on TCP connection")
			return
		}

		go c.handleConn(conn)
	}
}

func (c *Client) handleConn(conn net.Conn) {
	defer conn.Close()

	dialer := &websocket.Dialer{
		ReadBufferSize:   4 * 1024,
		WriteBufferSize:  4 * 1024,
		HandshakeTimeout: 10 * time.Second,
	}
	ws, _, err := dialer.Dial(c.WSUri, nil)
	if err != nil {
		log.Println("[ERR] shadowX: failed to dial websocket:", err)
		return
	}

	wsConn := NewConn(ws)
	defer wsConn.Close()

	c.proxyConn(conn, wsConn)
}

func (c *Client) proxyConn(conn net.Conn, rConn net.Conn) {
	closeCh := make(chan struct{}, 2)
	go c.copyConn(conn, rConn, closeCh)
	go c.copyConn(rConn, conn, closeCh)
	<-closeCh
}

func (c *Client) copyConn(dst net.Conn, src net.Conn, closeCh chan struct{}) {
	io.Copy(dst, src)
	closeCh <- struct{}{}
}
