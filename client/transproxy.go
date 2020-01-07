package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

type Transproxy struct {
	Port      uint16
	SocksPort uint16
}

func (t *Transproxy) Serve() (err error) {
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

func (t *Transproxy) handleConn(conn net.Conn) {
	dstHost, dstPort, conn, err := getOriginalDst(conn.(*net.TCPConn))
	if err != nil {
		log.Println("[ERR] transproxy: failed to get original destination host and port:", err)
		return
	}
	defer conn.Close()

	err = conn.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		log.Println("[ERR] transproxy: failed to set KeepAlive on TCP connection:", err)
		return
	}

	err = conn.(*net.TCPConn).SetKeepAlivePeriod(time.Hour)
	if err != nil {
		log.Println("[ERR] transproxy: failed to set KeepAlivePeriod on TCP connection:", err)
		return
	}

	dialSocksProxy, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", t.SocksPort), nil, proxy.Direct)
	if err != nil {
		log.Println("[ERR] transproxy: failed to create socks5 proxy:", err)
		return
	}

	rConn, err := dialSocksProxy.Dial("tcp", fmt.Sprintf("%s:%d", dstHost, dstPort))
	if err != nil {
		log.Println("[ERR] transproxy: failed to dail proxy connection:", err)
		return
	}
	defer rConn.Close()

	t.proxyConn(conn, rConn)
}

func (t *Transproxy) proxyConn(conn net.Conn, rConn net.Conn) {
	closeCh := make(chan struct{}, 2)
	go t.copyConn(conn, rConn, closeCh)
	go t.copyConn(rConn, conn, closeCh)
	<-closeCh
}

func (t *Transproxy) copyConn(dst net.Conn, src net.Conn, closeCh chan struct{}) {
	io.Copy(dst, src)
	closeCh <- struct{}{}
}

// get the original destination for the socket when redirect by linux iptables
// refer to https://raw.githubusercontent.com/missdeer/avege/master/src/inbound/redir/redir_iptables.go
//
const (
	SO_ORIGINAL_DST      = 80
	IP6T_SO_ORIGINAL_DST = 80
)

func getOriginalDst(clientConn *net.TCPConn) (host string, port uint16, newTCPConn *net.TCPConn, err error) {
	if clientConn == nil {
		err = errors.New("clientConn is nil")
		return
	}
	defer clientConn.Close()

	// test if the underlying fd is nil
	remoteAddr := clientConn.RemoteAddr()
	if remoteAddr == nil {
		err = errors.New("clientConn.fd is nil")
		return
	}

	// net.TCPConn.File() will cause the receiver's (clientConn) socket to be placed in blocking mode.
	// The workaround is to take the File returned by .File(), do getsockopt() to get the original
	// destination, then create a new *net.TCPConn by calling net.Conn.FileConn().  The new TCPConn
	// will be in non-blocking mode.  What a pain.
	clientConnFile, err := clientConn.File()
	if err != nil {
		return
	}
	defer clientConnFile.Close()

	// Get original destination
	// this is the only syscall in the Golang libs that I can find that returns 16 bytes
	// Example result: &{Multiaddr:[2 0 31 144 206 190 36 45 0 0 0 0 0 0 0 0] Interface:0}
	// port starts at the 3rd byte and is 2 bytes long (31 144 = port 8080)
	// IPv6 version, didn't find a way to detect network family
	// addr, err := syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IPV6, IP6T_SO_ORIGINAL_DST)
	// IPv4 address starts at the 5th byte, 4 bytes long (206 190 36 45)
	addr, err := syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		err = fmt.Errorf("syscall.GetsockoptIPv6Mreq: %v", err)
		return
	}

	newConn, err := net.FileConn(clientConnFile)
	if err != nil {
		return
	}

	newTCPConn, ok := newConn.(*net.TCPConn)
	if !ok {
		newConn.Close()
		err = fmt.Errorf("newConn is not a *net.TCPConn, instead it is: %T (%v)", newConn, newConn)
		return
	}

	host = fmt.Sprintf(
		"%d.%d.%d.%d",
		addr.Multiaddr[4],
		addr.Multiaddr[5],
		addr.Multiaddr[6],
		addr.Multiaddr[7],
	)
	port = uint16(addr.Multiaddr[2])<<8 + uint16(addr.Multiaddr[3])
	return
}
