package main

import (
	"flag"
	"fmt"
	"os"
	"ssx/client"
)

const version = "1.0.1"

func main() {
	wsURI := flag.String("ws", "ws://127.0.0.1/api/ssx", "server ws uri")
	socksPort := flag.Int("socks", 1086, "local socks proxy port")
	transporxyPort := flag.Int("transporxy", 0, "local tpc transparent proxy port")
	pacPort := flag.Int("pac", 0, "local pac server port")
	gfwlistURI := flag.String("gfwlist-uri", "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt", "gfwlist URI")
	printGFWList := flag.Bool("gfwlist", false, "print gfwlist")
	printVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *printVersion {
		fmt.Print(version)
		return
	}

	if *printGFWList {
		s := &client.PACServer{
			GFWListURI: *gfwlistURI,
		}
		list, err := s.GFWList(true)
		if err != nil {
			fmt.Println("can not download the gfwlist")
			os.Exit(1)
		}
		for host := range list {
			fmt.Println(host)
		}
		return
	}

	client.Run(*wsURI, uint16(*socksPort), uint16(*transporxyPort), uint16(*pacPort), *gfwlistURI)
}
