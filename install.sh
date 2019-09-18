#!/bin/sh

# stop shadowx firstly
enable=`dbus get shadowx_enable`
if [ "$enable" == "1" ];then
	restart=1
	dbus set shadowx_enable=0
	sh /koolshare/scripts/shadowx.sh
fi

# cp files
cp -rf /tmp/shadowx/bin/shadowx /koolshare/bin/shadowx
cp -rf /tmp/shadowx/res/icon-shadowx.png /koolshare/res/icon-shadowx.png
cp -rf /tmp/shadowx/scripts/shadowx.sh /koolshare/scripts/shadowx.sh
cp -rf /tmp/shadowx/webs/Module_shadowx.asp /koolshare/webs/Module_shadowx.asp

# delete install tar
rm -rf /tmp/shadowx* >/dev/null 2>&1

chmod a+x /koolshare/scripts/shadowx.sh
chmod 0755 /koolshare/bin/shadowx

CUR_VERSION=`/koolshare/bin/shadowx -v`
dbus remove shadowx_version
dbus set shadowx_version="$CUR_VERSION"
dbus set softcenter_module_shadowx_version="$CUR_VERSION"
dbus set softcenter_module_shadowx_title="Shadow X"
dbus set softcenter_module_shadowx_description="Break the GWF"

# re-enable shadowx
if [ "$restart" == "1" ];then
	dbus set shadowx_enable=1
	sh /koolshare/scripts/shadowx.sh
fi
