package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/url"
	"ssx/wsconn"
	"strings"

	"github.com/ije/gox/utils"
)

func ServeHttpProxy(conn *wsconn.Conn, debug bool) {
	defer conn.Close()

	var b [1024]byte
	n, err := conn.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == "443" {
		address = hostPortURL.Scheme + ":443"
	} else {
		if strings.Index(hostPortURL.Host, ":") == -1 {
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}

	targetConn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	if method == "CONNECT" {
		fmt.Fprint(conn, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		targetConn.Write(b[:n])
	}

	utils.ProxyConn(targetConn, conn, 0)
}
