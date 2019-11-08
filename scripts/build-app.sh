#!/bin/sh

platform=merlin_380
read -p "app platform (default is 'merlin_380')? " ap
if [ "$ap" != "" ]; then 
    platform="$ap"
fi
read -p "udpate configs ('yes' or 'no', default is 'no')? " updateConfigs

CUR_VERSION=`go run client-main.go -version`

export GOARCH=arm
export GOARM=5
export GOOS=linux


echo "compiling ssx-client(${GOOS}_$GOARCH)..."
rm -f ../app/$platform/bin/ssx
go build -o ../app/$platform/bin/ssx client-main.go
if [ "$?" != "0" ]; then 
    exit
fi

if [ "$updateConfigs" = "yes" ]; then
	echo "updating configs..."
	wget -O- 'http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest' | grep ipv4 | grep CN | awk -F\| '{ printf("%s/%d\n", $4, 32-log($5)/log(2)) }' > ../configs/chnroute.txt
	# todo: update gfwlist.conf/china-domains.txt
fi

cd ../app
mv $platform ssx
tar -czf ~/Downloads/ssx-${CUR_VERSION}.tar.gz ssx
mv ssx $platform
echo "~/Downloads/ssx-${CUR_VERSION}.tar.gz created"
