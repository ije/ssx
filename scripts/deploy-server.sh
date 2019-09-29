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
    scp -P $hostSSHPort supervisor.conf $loginUser@$host:/etc/supervisor/conf.d/ssx-server.conf
    if [ "$?" != "0" ]; then
        exit
    fi
fi

scp -P $hostSSHPort ssx-server $loginUser@$host:/tmp/ssx-server
if [ "$?" != "0" ]; then
    rm ssx-server
    exit
fi

echo "--- restart ssx-server..."
ssh -p $hostSSHPort $loginUser@$host << EOF
    supervisorctl status ssx-server
    if [ "$?" != "0" ]; then
        echo "error: no supervisor installed!"
        exit
    fi
    supervisorctl stop ssx-server
    rm -f /usr/local/bin/ssx-server
    mv -f /tmp/ssx-server /usr/local/bin/ssx-server
    chmod +x /usr/local/bin/ssx-server
    if [ "$init" = "yes" ]; then
        /usr/local/bin/ssx-server -i
        supervisorctl reload
    else
        supervisorctl start ssx-server
    fi
EOF

rm ssx-server
