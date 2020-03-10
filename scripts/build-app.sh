#!/bin/sh

platform=merlin_380
read -p "app platform (default is 'merlin_380')? " ap
if [ "$ap" != "" ]; then 
    platform="$ap"
fi
read -p "update configs ('yes' or 'no', default is 'no')? " updateConfigs

if [ "$updateConfigs" = "yes" ]; then
	echo "updating chnroute..."
	wget -O delegated-apnic-latest.txt 'http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest'
	cat delegated-apnic-latest.txt | grep ipv4 | grep CN | awk -F\| '{ printf("%s/%d\n", $4, 32-log($5)/log(2)) }' > ../app/$platform/configs/chnroute.txt

	# echo "updating gfwlist..."
	# gfwlist=`go run client-main.go -gfwlist`
	# echo "$gfwlist" > ../app/$platform/configs/gfwlist.conf
	
	# todo: update china-domains.txt
fi

CUR_VERSION=`go run client-main/main.go -version`

export GOARCH=arm
export GOARM=5
export GOOS=linux

echo "compiling ssx-client(${GOOS}_$GOARCH) v${CUR_VERSION}..."
rm -f ../app/$platform/bin/ssx
cd ./client-main
go build -o ../../app/$platform/bin/ssx main.go
if [ "$?" != "0" ]; then 
    exit
fi

cd ../../app
mv $platform ssx
tar -czf ~/Downloads/ssx-${CUR_VERSION}.tar.gz ssx
mv ssx $platform
echo "~/Downloads/ssx-${CUR_VERSION}.tar.gz created"
