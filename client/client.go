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

type Client struct {
	ServerURL string
	Port      uint16
}

func (c *Client) ServeUDP() (err error) {
	socket, err := net.ListenPacket("udp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return
	}
	defer socket.Close()

	dialer := &websocket.Dialer{
		ReadBufferSize:   16 * 1024,
		WriteBufferSize:  16 * 1024,
		HandshakeTimeout: 15 * time.Second,
	}
	ws, _, err := dialer.Dial(c.ServerURL, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	buffer := make([]byte, 16*1024)
	for {
		var n int
		var addr net.Addr
		n, addr, err = socket.ReadFrom(buffer)
		if err != nil {
			return
		}

		err = ws.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			return
		}

		var mt int
		var data []byte
		mt, data, err = ws.ReadMessage()
		if err != nil {
			return
		}

		if mt == websocket.BinaryMessage {
			_, err = socket.WriteTo(data, addr)
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) Serve() (err error) {
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

		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			err = tcpConn.SetKeepAlive(true)
			if err != nil {
				log.Println("client: failed to set KeepAlive on TCP connection")
				return
			}

			err = tcpConn.SetKeepAlivePeriod(time.Hour)
			if err != nil {
				log.Println("client: failed to set KeepAlivePeriod on TCP connection")
				return
			}
		}

		go c.handleConn(conn)
	}
}

func (c *Client) handleConn(conn net.Conn) (err error) {
	defer conn.Close()

	dialer := &websocket.Dialer{
		ReadBufferSize:   16 * 1024,
		WriteBufferSize:  16 * 1024,
		HandshakeTimeout: 15 * time.Second,
	}
	ws, _, err := dialer.Dial(c.ServerURL, nil)
	if err != nil {
		log.Println("client: failed to dail server:", err, c.ServerURL)
		return
	}

	wsConn := wsconn.New(ws)
	defer wsConn.Close()

	err = c.proxyConn(conn, wsConn)
	return
}

func (c *Client) proxyConn(conn net.Conn, rConn net.Conn) error {
	closeCh := make(chan error, 2)
	go c.copyConn(conn, rConn, closeCh)
	go c.copyConn(rConn, conn, closeCh)
	return <-closeCh
}

func (c *Client) copyConn(dst net.Conn, src net.Conn, closeCh chan error) {
	_, err := io.Copy(dst, src)
	closeCh <- err
}
