package client

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

type direct struct {
	conn net.Conn
}

func (d *direct) Dial(network, addr string) (net.Conn, error) {
	return d.conn, nil
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
		err = errors.New("ERR: clientConn is nil")
		return
	}

	// test if the underlying fd is nil
	remoteAddr := clientConn.RemoteAddr()
	if remoteAddr == nil {
		err = errors.New("ERR: clientConn.fd is nil")
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
	clientConn.Close()
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
		err = fmt.Errorf("ERR: newConn is not a *net.TCPConn, instead it is: %T (%v)", newConn, newConn)
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
