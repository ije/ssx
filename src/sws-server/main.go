package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"shadowx-server/server"
	"strings"

	"github.com/ije/gox/utils"

	"github.com/ije/rex"
)

const version = "0.2.0"

func main() {
	red := flag.String("redirect", "http://localhost", "redirect")
	d := flag.Bool("d", false, "debug mode")
	flag.Parse()

	httpClient := &http.Client{}
	handle := func(ctx *rex.Context) {
		if ctx.R.RequestURI == "/api/ws" {
			server.WS(ctx.W, ctx.R, *d)
			return
		}

		if strings.HasPrefix(*red, "file://") {
			ctx.File(strings.TrimPrefix(*red, "file://"))
			return
		}

		if strings.HasPrefix(*red, "http") {
			req, err := http.NewRequest("GET", *red+ctx.R.RequestURI, ctx.R.Body)
			if err != nil {
				ctx.End(500)
				return
			}

			for key, values := range ctx.R.Header {
				req.Header[key] = values
			}
			ip, _ := utils.SplitByLastByte(ctx.R.RemoteAddr, ':')
			if ctx.R.Header.Get("XX-Forwarded-For") != "" {
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

		ctx.Ok("sws server v" + version)
	}

	rex.Get("/*", handle)
	rex.Post("/*", handle)
	rex.Put("/*", handle)
	rex.Delete("/*", handle)

	if *d {
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
				CacheDir:  "/etc/shadowx/cert-cache",
			},
		},
	})
}
