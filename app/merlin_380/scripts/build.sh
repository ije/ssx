#!/bin/sh

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
# wget -O ../configs/accelerated-domains.china.conf https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf

cd ../../
mv merlin_380 ssx
tar -czf ~/Downloads/ssx.tar.gz ssx
mv ssx merlin_380
