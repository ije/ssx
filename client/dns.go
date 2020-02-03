package client

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type DNSServer struct {
	DohServer  string
	Port       uint16
	cache      sync.Map
	httpClient *http.Client
}

type DNSCache struct {
	Data    []byte
	Expires int64
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
			log.Println("[error] DNSServer: read upd packet:", err)
			continue
		}
		go s.proxyDNS(conn, addr, raw[:n])
	}
}

func (s *DNSServer) proxyDNS(conn *net.UDPConn, addr *net.UDPAddr, raw []byte) {
	var msg dnsmessage.Message
	err := msg.Unpack(raw)
	if err != nil {
		log.Println("[error] DNSServer: can not unpack dns:", err)
		return
	}
	packed, err := msg.Pack()
	if err != nil {
		log.Println("[error] DNSServer: can not pack dns:", err)
		return
	}

	query := base64.RawURLEncoding.EncodeToString(packed)
	v, ok := s.cache.Load(query)
	if ok {
		if c, ok := v.(DNSCache); ok {
			_, err = conn.WriteToUDP(c.Data, addr)
			if err != nil {
				log.Printf("[error] DNSServer: could not write to udp connection: %s", err)
				return
			}
			conn = nil
			if c.Expires > time.Now().Unix() {
				log.Printf("[debug] DNSServer: get data from cache: %s", query)
				return
			}
		}
	}

	ret, err := s.queryDNS(query)
	if err != nil {
		log.Printf("[error] DNSServer: %s", err)
		return
	}

	if conn != nil {
		_, err = conn.WriteToUDP(ret, addr)
		if err != nil {
			log.Printf("[error] DNSServer: could not write to udp connection: %s", err)
		}
	}

	s.cache.Store(query, DNSCache{
		Data:    ret,
		Expires: time.Now().Unix() + 3600,
	})
}

func (s *DNSServer) queryDNS(query string) (ret []byte, err error) {
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

	url := fmt.Sprintf("%s?dns=%s", s.DohServer, query)
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	r.Header.Set("Accept", "application/dns-message")
	r.Header.Set("Content-Type", "application/dns-message")

	resp, err := s.httpClient.Do(r)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		if len(ret) > 0 {
			err = fmt.Errorf("could not read message from response: %s - %s", string(ret), url)
		} else {
			err = fmt.Errorf("wrong response from DOH server got %s - %s", http.StatusText(resp.StatusCode), url)
		}
		ret = nil
	}
	return
}
