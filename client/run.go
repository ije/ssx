package client

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func Run(server string, ssl bool, socksPort uint16, transproxyPort uint16, dnsPort uint16, dohServer string, pacPort uint16, gfwlistURI string) {
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
			s := &DNSServer{
				DohServer: dohServer,
				Port:      dnsPort,
			}
			for {
				err := s.ServeDNS()
				if err != nil {
					log.Println("client(dns-proxy) shutdown:", err)
				}
				time.Sleep(time.Second / 10)
			}
		}()
		log.Println("dns proxy enable, doh server:", dohServer)
	}

	c := &SocksClient{
		ServerURL: strings.TrimRight(server, "/") + "/ssx/socks",
		SSL:       ssl,
		Port:      socksPort,
	}
	log.Println("start socks5 proxy, server:", server, "ssl:", ssl)
	for {
		err := c.Serve()
		if err != nil {
			log.Println("client(socks-proxy) shutdown:", err)
		}
		time.Sleep(time.Second / 10)
	}
}
