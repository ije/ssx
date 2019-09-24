#!/bin/bash

read -p "please enter deploy server hostname or ip: " host
if [ "$host" == "" ]; then
    exit 1;
fi

loginUser="root"
read -p "please enter the host ssh login user (default is 'root'): " user
if [ "$user" != "" ]; then
    loginUser="$user"
fi

hostSSHPort="22"
read -p "please enter the host ssh port (default is '22'): " port
if [ "$port" != "" ]; then
    hostSSHPort="$port"
fi

init="no"
read -p "initiate services ('yes' or 'no', default is 'no')? " ok
if [ "$ok" = "yes" ]; then
    init="yes"
fi

sh build-server.sh
if [ "$?" != "0" ]; then 
    exit
fi

echo "--- uploading..."
if [ "$init" = "yes" ]; then
    scp -P $hostSSHPort supervisor.conf $loginUser@$host:/etc/supervisor/conf.d/sws-server.conf
    if [ "$?" != "0" ]; then
        exit
    fi
fi

cd ../bin
scp -P $hostSSHPort sws-server $loginUser@$host:/tmp/sws-server
if [ "$?" != "0" ]; then
    rm sws-server
    exit
fi

echo "--- restart sws-server..."
ssh -p $hostSSHPort $loginUser@$host << EOF
    supervisorctl status sws-server
    if [ "$?" != "0" ]; then
        echo "error: no supervisor installed!"
        exit
    fi
    supervisorctl stop sws-server
    rm -f /usr/local/bin/sws-server
    mv -f /tmp/sws-server /usr/local/bin/sws-server
    chmod +x /usr/local/bin/sws-server
    if [ "$init" = "yes" ]; then
        /usr/local/bin/sws-server -i
        supervisorctl reload
    else
        supervisorctl start sws-server
    fi
EOF

rm sws-server
