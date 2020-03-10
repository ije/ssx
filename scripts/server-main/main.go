package main

import (
	"flag"
	"ssx/server"
)

func main() {
	redirect := flag.String("redirect", "http://localhost", "redirect")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	server.Serve(*redirect, *debug)
}
