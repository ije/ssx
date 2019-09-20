#!/bin/sh

export GOARCH=arm
export GOARM=5
export GOOS=linux

rm -f ../bin/shadowx
echo "--- compiling shadowx-client(${GOOS}_$GOARCH)..."
cd ../src/shadowx-client
go build -o ../../bin/shadowx main.go
cd ../../scripts

if [ ! -f "../bin/shadowx" ]; then
	exit 1
fi

echo "--- updating configs..."
# wget -O- 'http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest' | grep ipv4 | grep CN | awk -F\| '{ printf("%s/%d\n", $4, 32-log($5)/log(2)) }' > ../configs/chnroute.txt
# wget -O ../configs/accelerated-domains.china.conf https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf

cd ../../
tar -czf ~/Downloads/shadowx.tar.gz shadowx
