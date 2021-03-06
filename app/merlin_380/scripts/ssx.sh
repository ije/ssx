#!/bin/sh

# load path environment in dbus databse
eval `dbus export ssx`
source /koolshare/scripts/base.sh
platform=$(uname -m)
lan_ipaddr=$(nvram get lan_ipaddr)
server=`echo $ssx_server | grep -E "[a-zA-Z0-9\-\.]+"`
doh_server=`echo $ssx_doh_server | grep -E "https?://[a-zA-Z0-9\-\.]+.*"`

SOCKS5_PORT=1086
TPROXY_PORT=1087
PROXY_DNS_PORT=1053
LOCAL_DNS_PORT=5300
DNS=`echo $ssx_dns`
DNSMASQ_POSTCONF=/jffs/scripts/dnsmasq.postconf
SSX_SERVER_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/ssx-server.conf
CHINA_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/accelerated-domains.china.conf
GFWLIST_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/gfwlist.conf
BLACKLIST_DNSMASQ_CONFIG=/jffs/configs/dnsmasq.d/blacklist.conf

# check platform
if [ "$platform" != "armv7l" ]; then
    echo "[error] ssx can't run in '$platform', needs armv7l!"
    exit 1
fi

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
    ipset create gfwlist nethash >/dev/null 2>&1
    ipset create blacklist nethash >/dev/null 2>&1
    ipset flush chnroute >/dev/null 2>&1
    ipset flush gfwlist >/dev/null 2>&1
    ipset flush blacklist >/dev/null 2>&1

    sed -e "s/^/add chnroute &/g" /koolshare/configs/chnroute.txt | awk '{print $0} END{print "COMMIT"}' | ipset -R
}

apply_nat_rules() {
    iptables -t nat -N SSX

    # ignore server ip
    ip=`ping -c 1 $server | grep PING | awk -F\( '{print $2}' | awk -F\) '{print $1}'`
    iptables -t nat -A SSX -d $ip -j RETURN

    # ignore doh server ip
    doh_server_host="mozilla.cloudflare-dns.com"
    if [ -n "$doh_server" ]; then
        doh_server_host=`echo $doh_server | awk -F/ '{print $3}'`
    fi
    doh_server_ip=`ping -c 1 $doh_server_host | grep PING | awk -F\( '{print $2}' | awk -F\) '{print $1}'`
    if [ "$doh_server_ip" != "$ip" ]; then
        iptables -t nat -A SSX -d $doh_server_ip -j RETURN
    fi

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

    # allow connection to chinese IPs
    iptables -t nat -A SSX -p tcp -m set --match-set chnroute dst -j RETURN
    
    # force redirect blacklist ips
    iptables -t nat -A SSX -p tcp -m set --match-set blacklist dst -j REDIRECT --to-ports $TPROXY_PORT

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
    ipset destroy blacklist >/dev/null 2>&1

    # disable chromecast
    chromecast_nu=`iptables -t nat -L PREROUTING -v -n --line-numbers | grep "dpt:53" | awk '{print $1}'`
    if [ -n "$chromecast_nu" ]; then
        iptables -t nat -D PREROUTING $chromecast_nu >/dev/null 2>&1
    fi
}

create_dnsmasq_conf() {
    # dnsmasq postconf
    [ ! -L "$DNSMASQ_POSTCONF" ] && ln -sf /koolshare/configs/ssx_dnsmasq.postconf $DNSMASQ_POSTCONF
    
    # ssx servers
    echo "server=/$server/$DNS" >> $SSX_SERVER_DNSMASQ_CONFIG
    if [ -n "$doh_server" ]; then
        doh_server_host=`echo $doh_server | awk -F/ '{print $3}'`
        if [ "$doh_server_host" != "$server" ]; then
            echo "server=/$doh_server_host/$DNS" >> $SSX_SERVER_DNSMASQ_CONFIG
        fi
    else
        echo "server=/mozilla.cloudflare-dns.com/$DNS" >> $SSX_SERVER_DNSMASQ_CONFIG
    fi

    # china domains
    cat /koolshare/configs/china-domains.txt | sed "s/^/server=&\//g" | sed "s/$/\/&$DNS/g" | sort | awk '{if ($0!=line) print;line=$0}' >> $CHINA_DNSMASQ_CONFIG
    
    # gfwlist domains
    [ ! -L "$GFWLIST_DNSMASQ_CONFIG" ] && ln -sf /koolshare/configs/gfwlist.conf $GFWLIST_DNSMASQ_CONFIG

    # blacklist domains
    if [ -n "$ssx_blacklist_domains" ]; then
        echo "# blacklist domains" >> $BLACKLIST_DNSMASQ_CONFIG
        for t in $ssx_blacklist_domains
        do
            domain=`echo $t | sed 's/[ \t\r\n]*$//g' | sed "s/^www\.//gi" | grep -E "^[a-zA-Z0-9\-\.]+$"`
            if [ -n "$domain" ]; then
                echo "server=/.$domain/127.0.0.1#$PROXY_DNS_PORT" >> $BLACKLIST_DNSMASQ_CONFIG
                echo "ipset=/.$domain/blacklist" >> $BLACKLIST_DNSMASQ_CONFIG
            fi
        done
    fi
}

flush_dnsmasq_conf() {
    rm -f $DNSMASQ_POSTCONF
    rm -f $SSX_SERVER_DNSMASQ_CONFIG
    rm -f $CHINA_DNSMASQ_CONFIG
    rm -f $GFWLIST_DNSMASQ_CONFIG
    rm -f $BLACKLIST_DNSMASQ_CONFIG
}

start_ssx() {
    echo "starting SSX..."

    ssl=""
    if [ "$ssx_ssl" = "1" ]; then
        ssl="-ssl"
    fi
    dohServer=""
    if [ -n "$doh_server" ]; then
        dohServer="-doh-server $doh_server"
    fi
    /koolshare/bin/ssx -server $server $ssl -socks $SOCKS5_PORT -transporxy $TPROXY_PORT -dns $PROXY_DNS_PORT $dohServer >/dev/null 2>&1 &
    /koolshare/bin/chinadns -s $DNS,127.0.0.1:$PROXY_DNS_PORT -c /koolshare/configs/chnroute.txt -m -p $LOCAL_DNS_PORT >/dev/null 2>&1 &

    load_module
    create_ipset
    create_dnsmasq_conf
    service restart_dnsmasq
    apply_nat_rules
}

stop_ssx() {
    echo "stopping SSX..."

    killall ssx chinadns >/dev/null 2>&1

    flush_dnsmasq_conf
    service restart_dnsmasq
    flush_nat
}

case $ACTION in
start)
    if [ "$ssx_enable" = "1" -a -n "$server" ]; then
        start_ssx 
    fi
    ;;
stop)
    stop_ssx
    ;;
*)
    stop_ssx
    if [ "$ssx_enable" = "1" -a -n "$server" ]; then
        start_ssx
    fi
    ;;
esac
