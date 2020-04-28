package client

import (
	"fmt"
	"log"
	"net"
	"ssx/wsconn"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ije/gox/utils"
)

type ProxyClient struct {
	ServerURL string
	SSL       bool
	Port      uint16
}

func (c *ProxyClient) Serve() (err error) {
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

func (c *ProxyClient) handleConn(conn net.Conn) (err error) {
	defer conn.Close()

	dialer := &websocket.Dialer{
		ReadBufferSize:   8 * 1024,
		WriteBufferSize:  8 * 1024,
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

	err = utils.ProxyConn(conn, wsConn, 0)
	return
}
