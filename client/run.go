package client

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func Run(server string, socksPort uint16, transproxyPort uint16, dnsPort uint16, pacPort uint16, gfwlistURI string) {
	if transproxyPort > 0 {
		go func() {
			transproxy := &Transproxy{
				SocksPort: socksPort,
				Port:      transproxyPort,
			}
			for {
				transproxy.Serve()
				time.Sleep(time.Second / 10)
			}
		}()
		log.Println("transproxy enable")
	}

	if pacPort > 0 && len(gfwlistURI) > 0 {
		go func() {
			s := http.Server{
				Addr: fmt.Sprintf(":%d", pacPort),
				Handler: &PACServer{
					SocksPort:  socksPort,
					GFWListURI: gfwlistURI,
				},
			}
			for {
				s.ListenAndServe()
				time.Sleep(time.Second / 10)
			}
		}()
		log.Println("pac server enable")
	}

	if dnsPort > 0 {
		go func() {
			c := &Client{
				ServerURL: strings.TrimRight(server, "/") + "/api/ssx/dns",
				Port:      dnsPort,
			}
			for {
				err := c.ServeUDP()
				if err != nil {
					log.Println("client(dns-proxy) shutdown:", err)
				}
				time.Sleep(time.Second / 10)
			}
		}()
		log.Println("dns proxy enable")
	}

	clt := &Client{
		ServerURL: strings.TrimRight(server, "/") + "/api/ssx/socks",
		Port:      socksPort,
	}
	log.Println("start socks5 proxy:", clt.ServerURL)
	for {
		err := clt.Serve()
		if err != nil {
			log.Println("client(socks-proxy) shutdown:", err)
		}
		time.Sleep(time.Second / 10)
	}
}
