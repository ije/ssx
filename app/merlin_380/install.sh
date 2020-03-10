#!/bin/sh

# stop ssx firstly
enable=`dbus get ssx_enable`
if [ "$enable" == "1" ];then
	restart=1
	dbus set ssx_enable=0
	sh /koolshare/scripts/ssx.sh
fi

# copy files
cp -rf /tmp/ssx/bin/ssx /koolshare/bin/ssx
cp -rf /tmp/ssx/bin/chinadns /koolshare/bin/chinadns
cp -rf /tmp/ssx/res/icon-ssx.png /koolshare/res/icon-ssx.png
cp -rf /tmp/ssx/scripts/ssx.sh /koolshare/scripts/ssx.sh
cp -rf /tmp/ssx/webs/Module_ssx.asp /koolshare/webs/Module_ssx.asp
cp -rf /tmp/ssx/configs/chnroute.txt /koolshare/configs/chnroute.txt
cp -rf /tmp/ssx/configs/china-domains.txt /koolshare/configs/china-domains.txt
cp -rf /tmp/ssx/configs/gfwlist.conf /koolshare/configs/gfwlist.conf
cp -rf /tmp/ssx/configs/dnsmasq.postconf /koolshare/configs/ssx_dnsmasq.postconf

# delete install tar
rm -rf /tmp/ssx* >/dev/null 2>&1

chmod a+x /koolshare/scripts/ssx.sh
chmod a+x /koolshare/configs/ssx_dnsmasq.postconf
chmod 0755 /koolshare/bin/ssx
chmod 0755 /koolshare/bin/chinadns

CUR_VERSION=`/koolshare/bin/ssx -version`
dbus remove ssx_version
dbus set ssx_version="$CUR_VERSION"
dbus set softcenter_module_ssx_version="$CUR_VERSION"
dbus set softcenter_module_ssx_title="SSX"
dbus set softcenter_module_ssx_description="科学上网"

# try to create start_up script
if [ ! -L "/koolshare/init.d/S97ssx.sh" ]; then 
	ln -sf /koolshare/scripts/ssx.sh /koolshare/init.d/S97ssx.sh
fi

# re-enable ssx
if [ "$restart" == "1" ];then
	dbus set ssx_enable=1
	sh /koolshare/scripts/ssx.sh
fi
