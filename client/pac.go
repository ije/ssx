package client

import (
	"fmt"
	"log"
	"net/http"
)

type PACServer struct {
	SocksPort  uint16
	GFWListURI string
	gfwlist    map[string]struct{}
}

func (s *PACServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gfwlist, err := s.GFWList(false)
	if err != nil {
		http.Error(w, "can not get the gfwlist", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	pacTpl.Execute(w, map[string]interface{}{
		"gfwlist": gfwlist,
		"custom":  map[string]interface{}{},
		"proxy":   fmt.Sprintf("SOCKS5 127.0.0.1:%d", s.SocksPort),
	})
}

func (s *PACServer) GFWList(forceUpdate bool) (map[string]struct{}, error) {
	if s.gfwlist == nil {
		downloadedList, err := downloadAndParseGFWList(s.GFWListURI)
		if err != nil && !forceUpdate {
			log.Println("[warn] download GFWList:", err)
			downloadedList, err = parseGFWList([]byte(defaultGFWList))
		}
		if err != nil {
			return nil, err
		}
		s.gfwlist = downloadedList
	}
	return s.gfwlist, nil
}
