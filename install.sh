#!/bin/sh

# stop sws firstly
enable=`dbus get sws_enable`
if [ "$enable" == "1" ];then
	restart=1
	dbus set sws_enable=0
	sh /koolshare/scripts/sws.sh
fi

# cp files
cp -rf /tmp/sws/bin/sws /koolshare/bin/sws
cp -rf /tmp/sws/bin/chinadns /koolshare/bin/chinadns
cp -rf /tmp/sws/bin/dns2socks /koolshare/bin/dns2socks
cp -rf /tmp/sws/res/icon-sws.png /koolshare/res/icon-sws.png
cp -rf /tmp/sws/scripts/sws.sh /koolshare/scripts/sws.sh
cp -rf /tmp/sws/webs/Module_sws.asp /koolshare/webs/Module_sws.asp
cp -rf /tmp/sws/configs/chnroute.txt /koolshare/configs/chnroute.txt
cp -rf /tmp/sws/configs/china-domains.txt /koolshare/configs/china-domains.txt
cp -rf /tmp/sws/configs/gfwlist.conf /koolshare/configs/gfwlist.conf
cp -rf /tmp/sws/configs/dnsmasq.postconf /koolshare/configs/sws_dnsmasq.postconf

# delete install tar
rm -rf /tmp/sws* >/dev/null 2>&1

chmod a+x /koolshare/scripts/sws.sh
chmod a+x /koolshare/configs/sws_dnsmasq.postconf
chmod 0755 /koolshare/bin/sws
chmod 0755 /koolshare/bin/chinadns
chmod 0755 /koolshare/bin/dns2socks

CUR_VERSION=`/koolshare/bin/sws -v`
dbus remove sws_version
dbus set sws_version="$CUR_VERSION"
dbus set softcenter_module_sws_version="$CUR_VERSION"
dbus set softcenter_module_sws_title="Shadow WebSockets"
dbus set softcenter_module_sws_description="Breaking the GWF"

# re-enable sws
if [ "$restart" == "1" ];then
	dbus set sws_enable=1
	sh /koolshare/scripts/sws.sh
fi
