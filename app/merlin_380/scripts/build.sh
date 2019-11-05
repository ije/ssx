#!/bin/sh

CUR_VERSION=`go run client-main.go -version`

export GOARCH=arm
export GOARM=5
export GOOS=linux

rm -f ../bin/ssx
echo "--- compiling ssx-client(${GOOS}_$GOARCH)..."
go build -o ../bin/ssx client-main.go

if [ ! -f "../bin/ssx" ]; then
	exit 1
fi

echo "--- updating configs..."
# wget -O- 'http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest' | grep ipv4 | grep CN | awk -F\| '{ printf("%s/%d\n", $4, 32-log($5)/log(2)) }' > ../configs/chnroute.txt
 
cd ../../
mv merlin_380 ssx
tar -czf ~/Downloads/ssx-${CUR_VERSION}.tar.gz ssx
mv ssx merlin_380
echo "--- ~/Downloads/ssx-${CUR_VERSION}.tar.gz created"
