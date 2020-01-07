package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const dohServer = "https://mozilla.cloudflare-dns.com/dns-query"

func ServeDohProxy(w http.ResponseWriter, r *http.Request, debug bool) {
	dns := r.URL.Query().Get("dns")
	if dns == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	msg, header, err := queryDNS(dns)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	for key, values := range header {
		w.Header()[key] = values
	}
	w.Header().Set("server", "ssx-server")
	w.WriteHeader(200)
	w.Write(msg)
}

func queryDNS(dns string) (msg []byte, header http.Header, err error) {
	url := fmt.Sprintf("%s?dns=%s", dohServer, dns)
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	r.Header.Set("Accept", "application/dns-message")
	r.Header.Set("Content-Type", "application/dns-message")

	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(resp.Status)
		return
	}

	msg, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	header = resp.Header
	return
}
