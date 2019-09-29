#!/bin/bash

goos="linux"
read -p "please enter the deploy OS(default is 'linux'): " sys
if [ "$sys" != "" ]; then
	goos="$sys"
fi
export GOOS=$goos

goarch="amd64"
read -p "please enter the deploy OS Arch(default is 'amd64'): " arch
if [ "$arch" != "" ]; then
	goarch="$arch"
fi
export GOARCH=$goarch

echo "--- compiling the ssx-server(${goos}_$goarch)..."
go build -o ssx-server server-main.go
