package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"ssx/wsconn"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/ije/gox/utils"
	"github.com/ije/rex"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4 * 1024,
	WriteBufferSize: 4 * 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func upgrade(ctx *rex.Context) (conn *wsconn.Conn, err error) {
	ws, err := upgrader.Upgrade(ctx.W, ctx.R, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println("websocket:", err)
		}
		return
	}
	conn = wsconn.New(ws)
	return
}

func Serve(redirect string, debug bool) {
	httpClient := &http.Client{}
	handle := func(ctx *rex.Context) {
		if ctx.R.RequestURI == "/ssx/socks" {
			conn, err := upgrade(ctx)
			if err != nil {
				return
			}
			ServeSocks5Proxy(conn, debug)
			return
		}

		if ctx.R.RequestURI == "/ssx/http" {
			conn, err := upgrade(ctx)
			if err != nil {
				return
			}
			ServeHttpProxy(conn, debug)
			return
		}

		if strings.HasPrefix(ctx.R.RequestURI, "/ssx/dns-query") {
			ServeDohProxy(ctx.W, ctx.R, debug)
			return
		}

		if strings.HasPrefix(redirect, "file://") {
			ctx.File(strings.TrimPrefix(redirect, "file://"))
			return
		}

		if strings.HasPrefix(redirect, "http://") || strings.HasPrefix(redirect, "https://") {
			req, err := http.NewRequest(ctx.R.Method, redirect+ctx.R.RequestURI, ctx.R.Body)
			if err != nil {
				ctx.End(500, err.Error())
				return
			}

			for key, values := range ctx.R.Header {
				req.Header[key] = values
			}
			ip, _ := utils.SplitByLastByte(ctx.R.RemoteAddr, ':')
			if ctx.R.Header.Get("X-Forwarded-For") != "" {
				req.Header.Set("X-Forwarded-For", fmt.Sprintf("%s,%s", req.Header.Get("X-Forwarded-For"), ip))
			} else {
				req.Header.Set("X-Forwarded-For", ip)
			}
			if ctx.R.Header.Get("X-Real-IP") == "" {
				req.Header.Set("X-Real-IP", ip)
			}
			req.Header.Set("X-Real-Host", ctx.R.Host)
			resp, err := httpClient.Do(req)
			if err != nil {
				ctx.End(http.StatusBadGateway, err.Error())
				return
			}
			defer resp.Body.Close()

			for key, values := range resp.Header {
				ctx.W.Header()[key] = values
			}
			ctx.W.WriteHeader(resp.StatusCode)
			io.Copy(ctx.W, resp.Body)
			return
		}

		ctx.Ok("Hello world!")
	}

	rex.Get("/*", handle)
	rex.Post("/*", handle)
	rex.Put("/*", handle)
	rex.Patch("/*", handle)
	rex.Delete("/*", handle)

	if debug {
		rex.Serve(rex.ServerConfig{
			Port: 80,
		})
		return
	}

	rex.Serve(rex.ServerConfig{
		TLS: rex.TLSConfig{
			Port: 443,
			AutoTLS: rex.AutoTLSConfig{
				AcceptTOS: true,
				CacheDir:  "/etc/ssx/cert-cache",
			},
		},
	})
}
