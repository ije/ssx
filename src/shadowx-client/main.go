package main

import (
	"flag"
	"fmt"
	"shadowx-client/client"
)

const version = "0.1.0"

func main() {
	ws := flag.String("ws", "ws://localhost/api/ws", "server ws uri")
	sp := flag.Int("p", 1086, "local socks5 port")
	tp := flag.Int("tProxy", 1087, "local tpc proxy port")
	v := flag.Bool("v", false, "print shadowx client version")
	flag.Parse()

	if *v {
		fmt.Print(version)
		return
	}

	tProxy := &client.TProxy{
		SocksPort: uint16(*sp),
		Port:      uint16(*tp),
	}
	go tProxy.Serve()

	clt := &client.Client{
		WSUri: *ws,
		Port:  uint16(*sp),
	}
	err := clt.Serve()
	if err != nil {
		fmt.Println("client server shutdown", err)
	}
}
