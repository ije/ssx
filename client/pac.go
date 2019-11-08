package client

import (
	"fmt"
	"net/http"
)

type PACServer struct {
	SocksPort  uint16
	GFWListURI string
	gfwDomains map[string]struct{}
}

func (s *PACServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.gfwDomains == nil {
		domains, err := downloadAndParseGFWList(s.GFWListURI)
		if err != nil {
			http.Error(w, "internal server error", 500)
		}
		s.gfwDomains = domains
	}

	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	pacTpl.Execute(w, map[string]interface{}{
		"domains": s.gfwDomains,
		"custom":  map[string]int{},
		"proxy":   fmt.Sprintf("SOCKS5 127.0.0.1:%d", s.SocksPort),
	})
}
