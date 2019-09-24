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

echo "--- compiling the sms-server(${goos}_$goarch)..."
cd ../src/sws-server
go build -o ../../bin/sws-server main.go
