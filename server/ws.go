package server

import (
	"log"
	"net/http"
	"ssx/wsconn"

	"github.com/armon/go-socks5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  16 * 1024,
	WriteBufferSize: 16 * 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func serveWS(w http.ResponseWriter, r *http.Request, debug bool) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println("websocket:", err)
		}
		return
	}

	conn := wsconn.New(ws)
	defer conn.Close()

	socks, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Println("create socks5:", err)
		return
	}
	socks.ServeConn(conn)
}
