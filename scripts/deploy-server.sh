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
    scp -P $hostSSHPort supervisor.conf $loginUser@$host:/etc/supervisor/conf.d/shadowx-server.conf
    if [ "$?" != "0" ]; then
        exit
    fi
fi

cd ../bin
scp -P $hostSSHPort shadowx-server $loginUser@$host:/tmp/shadowx-server
if [ "$?" != "0" ]; then
    rm shadowx-server
    exit
fi

echo "--- restart shadowx-server..."
ssh -p $hostSSHPort $loginUser@$host << EOF
    supervisorctl status shadowx-server
    if [ "$?" != "0" ]; then
        echo "error: no supervisor installed!"
        exit
    fi
    supervisorctl stop shadowx-server
    rm -f /usr/local/bin/shadowx-server
    mv -f /tmp/shadowx-server /usr/local/bin/shadowx-server
    chmod +x /usr/local/bin/shadowx-server
    if [ "$init" = "yes" ]; then
        /usr/local/bin/shadowx-server -i
        supervisorctl reload
    else
        supervisorctl start shadowx-server
    fi
EOF

rm shadowx-server
