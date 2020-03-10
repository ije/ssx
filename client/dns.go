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

	"ssx/dns/dnsutil"
	"ssx/dns/response"

	"github.com/miekg/dns"
)

type DNSServer struct {
	DohServer  string
	Port       uint16
	cache      sync.Map
	httpClient *http.Client
}

func (s *DNSServer) ServeDNS() (err error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: int(s.Port)})
	if err != nil {
		return
	}
	defer conn.Close()

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

	for {
		buf := make([]byte, 512)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("[error] DNS: read upd packet:", err)
			continue
		}
		go s.proxyDNS(conn, addr, buf[:n])
	}
}

func (s *DNSServer) proxyDNS(conn *net.UDPConn, addr *net.UDPAddr, raw []byte) {
	m := new(dns.Msg)
	err := m.Unpack(raw)
	if err != nil {
		log.Println("[error] DNS: unpack dns message:", err)
		return
	}

	var do bool
	var ttl time.Duration

	now := time.Now().UTC()

	mt, opt := response.Typify(m, now)
	if opt != nil {
		do = opt.Do()
	}

	msgTTL := dnsutil.MinimalTTL(m, mt)
	if mt == response.ServerError {
		ttl = dnsutil.MinimalDefaultTTL
	} else if mt == response.NameError || mt == response.NoData {
		ttl = dnsutil.ComputeTTL(msgTTL, dnsutil.MinimalDefaultTTL, dnsutil.MaximumDefaulTTL/2)
	} else {
		ttl = dnsutil.ComputeTTL(msgTTL, 5*time.Minute, dnsutil.MaximumDefaulTTL)
	}

	hasKey, key := dnsutil.GetCacheKey(m, mt, do)
	if hasKey {
		v, ok := s.cache.Load(key)
		if ok {
			i := v.(*dnsutil.CacheItem)
			if i.TTL(now) > 0 {
				packed, _ := i.ToMsg(m, now).Pack()
				_, err = conn.WriteToUDP(packed, addr)
				if err != nil {
					log.Printf("[error] DNS: write upd packet: %v", err)
				}
				log.Printf("[debug] DNS: cache hit by %v", dns.Name(m.Question[0].Name))
				return
			}
			s.cache.Delete(key)
		}
	}

	log.Println("[debug] DNS:", dns.Name(m.Question[0].Name), key, mt, ttl)

	ret, err := s.queryDNS(m)
	if err != nil {
		log.Printf("[error] DNS: queryDNS: %s", err)
		return
	}

	respMsg := new(dns.Msg)
	err = respMsg.Unpack(ret)
	if err != nil {
		log.Printf("[error] DNS: parse response messsage: %s", err)
		return
	}

	i := dnsutil.NewCacheItem(respMsg, now, ttl)
	if hasKey {
		s.cache.Store(key, i)
	}

	packed, _ := i.ToMsg(m, now).Pack()
	_, err = conn.WriteToUDP(packed, addr)
	if err != nil {
		log.Printf("[error] DNS: write upd packet: %s", err)
	}
}

func (s *DNSServer) queryDNS(msg *dns.Msg) (ret []byte, err error) {
	packed, err := msg.Pack()
	if err != nil {
		return
	}

	query := base64.RawURLEncoding.EncodeToString(packed)
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
