package client

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type DNSServer struct {
	DohServer  string
	Port       uint16
	httpClient *http.Client
}

func (s *DNSServer) Serve() (err error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: int(s.Port)})
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var raw [512]byte
		n, addr, err := conn.ReadFromUDP(raw[:512])
		if err != nil {
			log.Println("[error] DNSServer read upd packet:", err)
			continue
		}
		go s.proxyDNS(conn, addr, raw[:n])
	}
}

func (s *DNSServer) proxyDNS(conn *net.UDPConn, addr *net.UDPAddr, raw []byte) {
	var msg dnsmessage.Message
	err := msg.Unpack(raw)
	if err != nil {
		log.Println("[error] DNSServer can not unpack dns:", err)
		return
	}
	packed, err := msg.Pack()
	if err != nil {
		log.Println("[error] DNSServer can not pack dns:", err)
		return
	}

	url := fmt.Sprintf("%s?dns=%s", s.DohServer, base64.RawURLEncoding.EncodeToString(packed))
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("[error] DNSServer could not create request: %s", err)
		return
	}
	r.Header.Set("Accept", "application/dns-message")
	// r.Header.Set("Content-Type", "application/dns-message")

	if s.httpClient == nil {
		s.httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 15 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 15 * time.Second,
			},
			Timeout: 15 * time.Second,
		}
	}

	resp, err := s.httpClient.Do(r)
	if err != nil {
		log.Printf("[error] DNSServer could not perform request: %s", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[error] DNSServer could not read message from response: %s", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		if len(body) > 0 {
			log.Printf("[error] DNSServer could not read message from response: %s - %s", string(body), url)
		} else {
			log.Printf("[error] DNSServer wrong response from DOH server got %s - %s", http.StatusText(resp.StatusCode), url)
		}
		return
	}

	_, err = conn.WriteToUDP(body, addr)
	if err != nil {
		log.Printf("[error] DNSServer could not write to udp connection: %s", err)
	}
}
