package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ije/gox/utils"
	"github.com/ije/rex"
)

func Serve(redirect string, debug bool) {
	httpClient := &http.Client{}
	handle := func(ctx *rex.Context) {
		if ctx.R.RequestURI == "/api/ssx/socks" {
			WebsocketToSocks5(ctx.W, ctx.R, debug)
			return
		}

		if ctx.R.RequestURI == "/api/ssx/dns" {
			WebsocketToDNS(ctx.W, ctx.R, debug)
			return
		}

		if strings.HasPrefix(redirect, "file://") {
			ctx.File(strings.TrimPrefix(redirect, "file://"))
			return
		}

		if strings.HasPrefix(redirect, "http") {
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
		rex.Serve(rex.Config{
			Port: 80,
		})
		return
	}

	rex.Serve(rex.Config{
		TLS: rex.TLSConfig{
			Port: 443,
			AutoTLS: rex.AutoTLSConfig{
				AcceptTOS: true,
				CacheDir:  "/etc/ssx/cert-cache",
			},
		},
	})
}
