#!/bin/sh

# load path environment in dbus databse
eval `dbus export ssx`
source /koolshare/scripts/base.sh
lan_ipaddr=$(nvram get lan_ipaddr)
ws_uri=`echo $ssx_ws_uri | grep -E "wss?://[a-zA-Z0-9\-\.]+.*"`

SOCKS5_PORT=1086
TPROXY_PORT=1087
DNS2SOCKS_PORT=1088
LOCAL_DNS_PORT=5300
DNS=`echo $ssx_dns`
DNSMASQ_POSTCONF=/jffs/scripts/dnsmasq.postconf
WS_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/ws.conf
CUSTOM_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/custom.conf
CHINA_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/china.conf
GFWLIST_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/gfwlist.conf

# check platform
case $(uname -m) in
    armv7l)
        echo "start installing ssx..."
    ;;
    *)
        echo "[error] ssx can't run in \"$(uname -m)\", needs koolshare merlin armv7l!"
        exit 1
    ;;
esac

get_lan_cidr() {
	netmask=`nvram get lan_netmask`
	local x=${netmask##*255.}
	set -- 0^^^128^192^224^240^248^252^254^ $(( (${#netmask} - ${#x})*2 )) ${x%%.*}
	x=${1%%$3*}
	suffix=$(( $2 + (${#x}/4) ))
	echo $lan_ipaddr/$suffix
}

load_module() {
	xt=`lsmod | grep xt_set`
	OS=$(uname -r)
	if [ -f /lib/modules/${OS}/kernel/net/netfilter/xt_set.ko ] && [ -z "$xt" ];then
		insmod /lib/modules/${OS}/kernel/net/netfilter/xt_set.ko
	fi
}

create_ipset() {
    ipset create chnroute nethash >/dev/null 2>&1
    ipset create gfwlist iphash >/dev/null 2>&1
    ipset create custom iphash >/dev/null 2>&1
    ipset flush chnroute >/dev/null 2>&1
    ipset flush gfwlist >/dev/null 2>&1
    ipset flush custom >/dev/null 2>&1

    sed -e "s/^/add chnroute &/g" /koolshare/configs/chnroute.txt | awk '{print $0} END{print "COMMIT"}' | ipset -R
}

apply_nat_rules() {
    iptables -t nat -N SSX

    # ignore server ip
    host=`echo $ws_uri | awk -F/ '{print $3}'`
    ip=`ping -c 1 $host | grep PING | awk -F\( '{print $2}' | awk -F\) '{print $1}'`
    iptables -t nat -A SSX -d $ip -j RETURN

    # ignore internal ips
    iptables -t nat -A SSX -d 0.0.0.0/8 -j RETURN
    iptables -t nat -A SSX -d 10.0.0.0/8 -j RETURN
    iptables -t nat -A SSX -d 100.64.0.0/10 -j RETURN
    iptables -t nat -A SSX -d 127.0.0.0/8 -j RETURN
    iptables -t nat -A SSX -d 169.254.0.0/16 -j RETURN
    iptables -t nat -A SSX -d 172.16.0.0/12 -j RETURN
    iptables -t nat -A SSX -d 192.168.0.0/16 -j RETURN
    iptables -t nat -A SSX -d 224.0.0.0/4 -j RETURN
    iptables -t nat -A SSX -d 240.0.0.0/4 -j RETURN

    # force redirect custom ips
    iptables -t nat -A SSX -p tcp -m set --match-set custom dst -j REDIRECT --to-ports $TPROXY_PORT

    # allow connection to chinese IPs
    iptables -t nat -A SSX -p tcp -m set --match-set chnroute dst -j RETURN
    
    # redirect gfwlist ips
    iptables -t nat -A SSX -p tcp -m set --match-set gfwlist dst -j REDIRECT --to-ports $TPROXY_PORT

    # apply chain to table
    iptables -t nat -I PREROUTING -p tcp -j SSX

    # enable chromecast
    chromecast_nu=`iptables -t nat -L PREROUTING -v -n --line-numbers | grep "dpt:53" | awk '{print $1}'`
	if [ -z "$chromecast_nu" ]; then
        iptables -t nat -A PREROUTING -p udp -s $(get_lan_cidr) --dport 53 -j DNAT --to $lan_ipaddr >/dev/null 2>&1
    fi
}

flush_nat() {
    # flush rules and set if any
	nat_indexs=`iptables -nvL PREROUTING -t nat | sed 1,2d | sed -n '/SSX/=' | sort -r`
	for nat_index in $nat_indexs
	do
        iptables -t nat -D PREROUTING $nat_index >/dev/null 2>&1
	done

    iptables -t nat -F SSX >/dev/null 2>&1
    iptables -t nat -X SSX >/dev/null 2>&1

    ipset destroy chnroute >/dev/null 2>&1
    ipset destroy gfwlist >/dev/null 2>&1
    ipset destroy custom >/dev/null 2>&1
    
    # disable chromecast
    chromecast_nu=`iptables -t nat -L PREROUTING -v -n --line-numbers | grep "dpt:53" | awk '{print $1}'`
	if [ -n "$chromecast_nu" ]; then
        iptables -t nat -D PREROUTING $chromecast_nu >/dev/null 2>&1
    fi
}

create_dnsmasq_conf() {
    [ ! -L "$DNSMASQ_POSTCONF" ] && ln -sf /koolshare/configs/ssx_dnsmasq.postconf $DNSMASQ_POSTCONF
    ws_host=`echo $ws_uri | awk -F/ '{print $3}'`
    echo "server=/$ws_host/$DNS" > $WS_DNSMASQ_CONFIG
    cat /koolshare/configs/china-domains.txt | sed "s/^/server=&\/./g" | sed "s/$/\/&$DNS/g" | sort | awk '{if ($0!=line) print;line=$0}' >> $CHINA_DNSMASQ_CONFIG
    [ ! -L "$GFWLIST_DNSMASQ_CONFIG" ] && ln -sf /koolshare/configs/gfwlist.conf $GFWLIST_DNSMASQ_CONFIG
    if [ -n "$ssx_custom_domains" ]; then
        echo "$ssx_custom_domains" | sed "s/^/server=&\/./g" | sed "s/$/\/127\.0\.0\.1#&$LOCAL_DNS_PORT/g" | sort | awk '{if ($0!=line) print;line=$0}' >> $CUSTOM_DNSMASQ_CONFIG
        echo "$ssx_custom_domains" | sed "s/^/ipset=&\/./g" | sed "s/$/\/custom/g" | sort | awk '{if ($0!=line) print;line=$0}' >> $CUSTOM_DNSMASQ_CONFIG
    fi
}

flush_dnsmasq_conf() {
    rm -f $DNSMASQ_POSTCONF
    rm -f $WS_DNSMASQ_CONFIG
    rm -f $CHINA_DNSMASQ_CONFIG
    rm -f $GFWLIST_DNSMASQ_CONFIG
    rm -f $CUSTOM_DNSMASQ_CONFIG
}

start_ssx() {
    echo "starting shadow X..."

    /koolshare/bin/ssx -ws $ws_uri -socks $SOCKS5_PORT -transporxy $TPROXY_PORT >/dev/null 2>&1 &
    /koolshare/bin/dns2socks 127.0.0.1:$SOCKS5_PORT 8.8.8.8:53 127.0.0.1:$DNS2SOCKS_PORT >/dev/null 2>&1 &
    /koolshare/bin/chinadns -s $DNS,127.0.0.1:$DNS2SOCKS_PORT -c /koolshare/configs/chnroute.txt -m -p $LOCAL_DNS_PORT >/dev/null 2>&1 &

    load_module
    create_ipset
    create_dnsmasq_conf
    service restart_dnsmasq
    apply_nat_rules

    # try to create start_up file
    if [ ! -L "/koolshare/init.d/S97Shadowx.sh" ]; then 
        ln -sf /koolshare/scripts/ssx.sh /koolshare/init.d/S97Shadowx.sh
    fi
}

stop_ssx() {
    echo "stopping shadow X..."

    killall ssx chinadns dns2socks >/dev/null 2>&1

    flush_dnsmasq_conf
    service restart_dnsmasq
    flush_nat
}

case $ACTION in
start)
    if [ "$ssx_enable" = "1" -a -n "$ws_uri" ]; then
        start_ssx 
    fi
    ;;
stop)
    stop_ssx
    ;;
*)
    stop_ssx
    if [ "$ssx_enable" = "1" -a -n "$ws_uri" ]; then
        start_ssx
    fi
    ;;
esac
