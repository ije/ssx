package main

import (
	"flag"
	"fmt"
)

const version = "0.0.1"

func main() {
	v := flag.Bool("v", false, "print shadowx version")
	flag.Parse()

	if *v {
		fmt.Print(version)
		return
	}

	fmt.Println("hello world!")
}
