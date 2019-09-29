package client

import (
	"fmt"
	"net/http"
)

func Run(wsURI string, socksPort uint16, transproxyPort uint16, pacPort uint16, gfwlistURI string) {
	if transproxyPort > 0 {
		transproxy := &Transproxy{
			SocksPort: socksPort,
			Port:      transproxyPort,
		}
		go transproxy.Serve()
	}

	if pacPort > 0 && len(gfwlistURI) > 0 {
		s := http.Server{
			Addr: fmt.Sprintf(":%d", pacPort),
			Handler: &PACServer{
				SocksPort:  socksPort,
				GFWListURI: gfwlistURI,
			},
		}
		go s.ListenAndServe()
	}

	clt := &Client{
		WSUri: wsURI,
		Port:  socksPort,
	}
	err := clt.Serve()
	if err != nil {
		fmt.Println("client server shutdown:", err)
	}
}
