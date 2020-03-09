package client

import (
	"fmt"
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
	if forceUpdate {
		downloadedList, err := downloadAndParseGFWList(s.GFWListURI)
		if err != nil {
			return nil, fmt.Errorf("can not download/parse the GFWList: %v", err)
		}
		s.gfwlist = downloadedList
	}

	if s.gfwlist == nil {
		var err error
		s.gfwlist, err = parseGFWList([]byte(defaultGFWList))
		if err != nil {
			panic(err)
		}
	}
	return s.gfwlist, nil
}
