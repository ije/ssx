#!/bin/sh
# load path environment in dbus databse
eval `dbus export shadowx`
source /koolshare/scripts/base.sh
DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/shadowx.conf 

# check platform
case $(uname -m) in
    armv7l)
        echo "start installing shadowx..."
    ;;
    *)
        echo "[error] shadowx can't run in \"$(uname -m)\", needs koolshare merlin armv7l!"
        exit 1
    ;;
esac

start_shadowx(){
    /koolshare/bin/shadowx
    echo "" > $DNSMASQ_CONFIG
    service restart_dnsmasq

    # try to create start_up file
    if [ ! -L "/koolshare/init.d/S97Shadowx.sh" ]; then 
        ln -sf /koolshare/scripts/shadowx.sh /koolshare/init.d/S97Shadowx.sh
    fi
}

stop_shadowx(){
    # clear start up command line in firewall-start
    killall shadowx
    rm -f $DNSMASQ_CONFIG
    service restart_dnsmasq
}

case $ACTION in
start)
    if [ "$shadowx_enable" == "1" ]; then
        start_shadowx 
    fi
    ;;
stop)
    close_port >/dev/null 2>&1
    stop_shadowx
    ;;
*)
    stop_shadowx
    if [ "$shadowx_enable" == "1" ]; then
        start_shadowx
    fi
    ;;
esac
