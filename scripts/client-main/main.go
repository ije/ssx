package main

import (
	"flag"
	"fmt"
	"os"

	"ssx/client"
)

const version = "1.2.4"

func main() {
	server := flag.String("server", "127.0.0.1", "server address")
	ssl := flag.Bool("ssl", false, "use ssl connection")
	socksPort := flag.Int("socks-port", 1086, "socks proxy port")
	httpproxyPort := flag.Int("httpproxy-port", 0, "http proxy port")
	transporxyPort := flag.Int("transporxy-proxy", 0, "tpc transparent proxy port")
	pacPort := flag.Int("pac-port", 0, "pac server port")
	dnsPort := flag.Int("dns-port", 0, "dns proxy port")
	dohServer := flag.String("doh-server", "https://mozilla.cloudflare-dns.com/dns-query", "doh server")
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

	client.Run(*server, *ssl, uint16(*socksPort), uint16(*httpproxyPort), uint16(*transporxyPort), uint16(*dnsPort), *dohServer, uint16(*pacPort), *gfwlistURI)
}
