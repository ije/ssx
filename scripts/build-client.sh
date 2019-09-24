#!/bin/sh

export GOARCH=arm
export GOARM=5
export GOOS=linux

rm -f ../bin/sws
echo "--- compiling sws-client(${GOOS}_$GOARCH)..."
cd ../src/sws-client
go build -o ../../bin/sws main.go
cd ../../scripts

if [ ! -f "../bin/sws" ]; then
	exit 1
fi

echo "--- updating configs..."
# wget -O- 'http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest' | grep ipv4 | grep CN | awk -F\| '{ printf("%s/%d\n", $4, 32-log($5)/log(2)) }' > ../configs/chnroute.txt
# wget -O ../configs/accelerated-domains.china.conf https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf

cd ../../
tar -czf ~/Downloads/sws.tar.gz sws
