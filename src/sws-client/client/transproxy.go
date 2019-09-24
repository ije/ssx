package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/net/proxy"
)

type TransProxy struct {
	Port      uint16
	SocksPort uint16
}

func (t *TransProxy) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))
	if err != nil {
		return
	}

	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			return
		}

		go t.handleConn(conn)
	}
}

func (t *TransProxy) handleConn(conn net.Conn) {
	dstHost, dstPort, conn, err := getOriginalDst(conn.(*net.TCPConn))
	if err != nil {
		log.Println("[ERR] shadowX: failed to get original destination host and port:", err)
		return
	}
	defer conn.Close()
	

	err = conn.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		log.Println("[ERR] shadowX: failed to set KeepAlive on TCP connection:", err)
		return
	}

	err = conn.(*net.TCPConn).SetKeepAlivePeriod(time.Hour)
	if err != nil {
		log.Println("[ERR] shadowX: failed to set KeepAlivePeriod on TCP connection:", err)
		return
	}

	dialSocksProxy, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", t.SocksPort), nil, proxy.Direct)
	if err != nil {
		log.Println("[ERR] shadowX: failed to create socks5 proxy:", err)
		return
	}

	rConn, err := dialSocksProxy.Dial("tcp", fmt.Sprintf("%s:%d", dstHost, dstPort))
	if err != nil {
		log.Println("[ERR] shadowX: failed to dail proxy connection:", err)
		return
	}
	defer rConn.Close()

	t.proxyConn(conn, rConn)
}

func (t *TransProxy) proxyConn(conn net.Conn, rConn net.Conn) {
	closeCh := make(chan struct{}, 2)
	go t.copyConn(conn, rConn, closeCh)
	go t.copyConn(rConn, conn, closeCh)
	<-closeCh
}

func (t *TransProxy) copyConn(dst net.Conn, src net.Conn, closeCh chan struct{}) {
	io.Copy(dst, src)
	closeCh <- struct{}{}
}
