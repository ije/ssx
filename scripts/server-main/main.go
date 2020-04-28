package main

import (
	"flag"
	"fmt"

	"ssx/server"
)

const version = "1.2.4"

func main() {
	redirect := flag.String("redirect", "http://localhost", "redirect")
	debug := flag.Bool("debug", false, "debug mode")
	printVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *printVersion {
		fmt.Print(version)
		return
	}

	server.Serve(*redirect, *debug)
}
