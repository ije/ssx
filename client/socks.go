package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"ssx/wsconn"
	"time"

	"github.com/gorilla/websocket"
)

type SocksClient struct {
	ServerURL string
	SSL       bool
	Port      uint16
}

func (c *SocksClient) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return
	}
	defer l.Close()

	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			return
		}

		go c.handleConn(conn)
	}
}

func (c *SocksClient) handleConn(conn net.Conn) (err error) {
	defer conn.Close()

	dialer := &websocket.Dialer{
		ReadBufferSize:   4 * 1024,
		WriteBufferSize:  4 * 1024,
		HandshakeTimeout: 15 * time.Second,
	}
	proto := "ws"
	if c.SSL {
		proto = "wss"
	}
	ws, _, err := dialer.Dial(proto+"://"+c.ServerURL, nil)
	if err != nil {
		log.Println("client: failed to dail server:", err, proto+"://"+c.ServerURL)
		return
	}

	wsConn := wsconn.New(ws)
	defer wsConn.Close()

	err = c.proxyConn(conn, wsConn)
	return
}

func (c *SocksClient) proxyConn(conn net.Conn, rConn net.Conn) error {
	closeCh := make(chan error, 2)
	go c.copyConn(conn, rConn, closeCh)
	go c.copyConn(rConn, conn, closeCh)
	return <-closeCh
}

func (c *SocksClient) copyConn(dst net.Conn, src net.Conn, closeCh chan error) {
	_, err := io.Copy(dst, src)
	closeCh <- err
}
