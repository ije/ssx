package client

import (
	"fmt"
	"net/http"
)

type PACServer struct {
	SocksPort  uint16
	GFWListURI string
	gfwList    map[string]struct{}
}

func (s *PACServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gfwList, err := s.GFWList()
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	pacTpl.Execute(w, map[string]interface{}{
		"gfwlist": gfwList,
		"custom":  map[string]int{},
		"proxy":   fmt.Sprintf("SOCKS5 127.0.0.1:%d", s.SocksPort),
	})
}

func (s *PACServer) GFWList() (map[string]struct{}, error) {
	if s.gfwList == nil {
		Hosts, err := downloadAndParseGFWList(s.GFWListURI)
		if err != nil {
			return nil, err
		}
		s.gfwList = Hosts
	}
	return s.gfwList, nil
}
