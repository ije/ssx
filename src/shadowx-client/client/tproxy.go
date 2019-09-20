package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type TProxy struct {
	Port      uint16
	SocksPort uint16
}

func (t *TProxy) Serve() (err error) {
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

func (t *TProxy) handleConn(conn net.Conn) {
	dstHost, dstPort, conn, err := getOriginalDst(conn.(*net.TCPConn))
	if err != nil {
		log.Println("[ERR] shadowX: failed to get original destination host and port:", err)
		return
	}
	defer conn.Close()

	log.Println("[dstHost, dstPort]", dstHost, dstPort)

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

func (t *TProxy) proxyConn(conn net.Conn, rConn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go t.copyConn(rConn, conn, &wg)
	go t.copyConn(conn, rConn, &wg)
	wg.Wait()
}

func (t *TProxy) copyConn(dst net.Conn, src net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	io.Copy(dst, src)
}
