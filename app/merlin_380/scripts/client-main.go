package main

import (
	"flag"
	"fmt"
	"ssx/client"
)

const version = "0.3.1"

func main() {
	wsURI := flag.String("ws", "ws://127.0.0.1/api/ws", "server ws uri")
	socksPort := flag.Int("socks", 1086, "local socks proxy port")
	transporxyPort := flag.Int("transporxy", 0, "local tpc transparent proxy port")
	pacPort := flag.Int("pac", 0, "local pac server port")
	gfwlistURI := flag.String("gfwlist", "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt", "gfwlist URI")
	printVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *printVersion {
		fmt.Print(version)
		return
	}

	client.Run(*wsURI, uint16(*socksPort), uint16(*transporxyPort), uint16(*pacPort), *gfwlistURI)
}
