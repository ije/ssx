package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

func WebsocketToDNS(w http.ResponseWriter, r *http.Request, debug bool) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println("websocket:", err)
		}
		return
	}
	defer ws.Close()

	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(8, 8, 8, 8),
		Port: 53,
	})
	if err != nil {
		log.Println("net:", err)
		return
	}
	defer socket.Close()

	buffer := make([]byte, 16*1024)
	for {
		mt, data, err := ws.ReadMessage()
		if err != nil {
			log.Println("websocket.ReadMessage:", err)
			return
		}

		if mt != websocket.BinaryMessage {
			continue
		}

		_, err = socket.Write(data)
		if err != nil {
			fmt.Println("udp.write:", err)
			return
		}

		var n int
		n, _, err = socket.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("udp.read:", err)
			return
		}

		err = ws.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			log.Println("websocket.WriteMessage:", err)
			return
		}
	}

}
