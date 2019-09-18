#!/bin/sh

export GOARCH=arm
export GOARM=5
export GOOS=linux

rm -f ../bin/shadowx
echo "--- compiling the shadowx(${GOOS}_$GOARCH)..."
go build -o ../bin/shadowx ../src/shadowx/main.go

if [ ! -f "../bin/shadowx" ]; then
	exit 1
fi

cd ../../
tar -czf ~/Downloads/shadowx.tar.gz shadowx