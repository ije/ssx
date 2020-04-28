package client

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func Run(server string, ssl bool, socksPort uint16, httpProxyPort uint16, transproxyPort uint16, dnsPort uint16, dohServer string, pacPort uint16, gfwlistURI string) {
	if transproxyPort > 0 {
		go func() {
			transproxy := &Transproxy{
				SocksPort: socksPort,
				Port:      transproxyPort,
			}
			for {
				transproxy.Serve()
				time.Sleep(time.Second)
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
				time.Sleep(time.Second)
			}
		}()
		log.Printf("pac server enable: http://localhost:%d/proxy.pac", pacPort)
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
				time.Sleep(time.Second)
			}
		}()
		log.Println("dns proxy enable, doh by", dohServer)
	}

	if httpProxyPort > 0 {
		go func() {
			c := &ProxyClient{
				ServerURL: strings.TrimRight(server, "/") + "/ssx/http",
				SSL:       ssl,
				Port:      httpProxyPort,
			}
			for {
				err := c.Serve()
				if err != nil {
					log.Println("client(http-proxy) shutdown:", err)
				}
				time.Sleep(time.Second)
			}
		}()
		log.Printf("http proxy enable: http://localhost:%d", httpProxyPort)
	}

	c := &ProxyClient{
		ServerURL: strings.TrimRight(server, "/") + "/ssx/socks",
		SSL:       ssl,
		Port:      socksPort,
	}
	log.Printf("start socks5 proxy, server: %s, ssl: %v", server, ssl)
	for {
		err := c.Serve()
		if err != nil {
			log.Println("client(socks-proxy) shutdown:", err)
		}
		time.Sleep(time.Second)
	}
}
