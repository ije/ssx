package client

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func downloadAndParseGFWList(uri string) (domains map[string]struct{}, err error) {
	res, err := http.Get(uri)
	if err != nil {
		return
	}
	defer res.Body.Close()

	domains = map[string]struct{}{}

	decoder := base64.NewDecoder(base64.StdEncoding, res.Body)
	reader := bufio.NewReader(decoder)
	for {
		var line []byte
		line, _, err = reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		domain := parse(string(line))
		if domain != "" {
			domains[domain] = struct{}{}
		}
	}
	return
}

/*
parse autoproxy gfwlist line,
return domain name
*/
func parse(line string) string {

	/* remove space */
	line = strings.TrimSpace(line)

	if line == "" {
		return ""
	}

	/* ignore ip address */
	if net.ParseIP(line) != nil {
		return ""
	}

	/* ignore pattern */
	if strings.Index(line, ".") == -1 {
		return ""
	}

	/* ignore comment, whitelist, regex */
	if line[0] == '[' ||
		line[0] == '!' ||
		line[0] == '/' ||
		line[0] == '@' {
		return ""
	}

	return getHostname(line)
}

func getHostname(line string) string {
	c := line[0]
	ss := line

	/* replace '*' */
	if strings.Index(ss, "/") == -1 {
		if strings.Index(ss, "*") != -1 && ss[:2] != "||" {
			ss = strings.Replace(ss, "*", "/", -1)
		}
	}

	switch c {
	case '.':
		ss = fmt.Sprintf("http://%s", ss[1:])
	case '|':
		switch ss[1] {
		case '|':
			ss = fmt.Sprintf("http://%s", ss[2:])
		default:
			ss = ss[1:]
		}
	default:
		if !strings.HasPrefix(ss, "http") {
			ss = fmt.Sprintf("http://%s", ss)
		}
	}

	/* process */
	u, err := url.Parse(ss)
	if err != nil {
		log.Printf("%s: %s\n", line, err)
		return ""
	}
	host := u.Host
	if n := strings.Index(host, "*"); n != -1 {
		for i := n; i < len(host); i++ {
			if host[i] == '.' {
				host = host[i:]
				break
			}
		}
	}
	return strings.TrimLeft(host, ".")
}
