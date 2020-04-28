package server

import (
	"log"
	"ssx/wsconn"

	"github.com/armon/go-socks5"
)

func ServeSocks5Proxy(conn *wsconn.Conn, debug bool) {
	defer conn.Close()

	socks, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Println("create socks5:", err)
		return
	}
	socks.ServeConn(conn)
}
