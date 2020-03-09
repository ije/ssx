package client

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/ije/gox/cache"
	"golang.org/x/net/dns/dnsmessage"
)

type DNSServer struct {
	DohServer  string
	Port       uint16
	cache      cache.Cache
	httpClient *http.Client
}

func (s *DNSServer) Serve() (err error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: int(s.Port)})
	if err != nil {
		return
	}
	defer conn.Close()

	if s.cache == nil {
		s.cache, err = cache.New("memory?gcInterval=1h")
		if err != nil {
			return
		}
	}

	for {
		buf := make([]byte, 512)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("[error] DNSServer: read upd packet:", err)
			continue
		}
		go s.proxyDNS(conn, addr, buf[:n])
	}
}

func (s *DNSServer) proxyDNS(conn *net.UDPConn, addr *net.UDPAddr, raw []byte) {
	var msg dnsmessage.Message
	err := msg.Unpack(raw)
	if err != nil {
		log.Println("[error] DNSServer: unpack dns message:", err)
		return
	}
	packed, err := msg.Pack()
	if err != nil {
		log.Println("[error] DNSServer: pack dns:", err)
		return
	}

	query := base64.RawURLEncoding.EncodeToString(packed)

	cachedRet, err := s.cache.Get(query)
	if err == nil {
		_, err = conn.WriteToUDP(cachedRet, addr)
		if err != nil {
			log.Printf("[error] DNSServer: write upd packet: %s", err)
		}
		log.Printf("[debug] DNSServer: cache hit by %s", query)
		return
	}

	ret, err := s.queryDNS(query)
	if err != nil {
		log.Printf("[error] DNSServer: queryDNS: %s", err)
		return
	}
	s.cache.SetTemp(query, ret, time.Hour)

	_, err = conn.WriteToUDP(ret, addr)
	if err != nil {
		log.Printf("[error] DNSServer: write upd packet: %s", err)
	}
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
