package server

import (
	"log"
	"net/http"

	"shadowx-server/external/socks5"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  16 * 1024,
	WriteBufferSize: 16 * 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Serve(w http.ResponseWriter, r *http.Request, debug bool) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println("websocket:", err)
		}
		return
	}

	conn := NewConn(ws)
	defer conn.Close()

	s5, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Println("socks5.New", err)
		return
	}
	s5.ServeConn(conn)
}
