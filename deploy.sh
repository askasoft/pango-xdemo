#!/bin/bash -ex

APPNAME=xdemo
APPHOME=/app/${APPNAME}
PREFIX=

if [ -z "${LOG_LEVEL}" ]; then
	LOG_LEVEL=INFO
fi


mkdir -p ${APPHOME}/conf
rm -rf ${APPHOME}/${APPNAME}
rm -rf ${APPHOME}/tpls
rm -rf ${APPHOME}/txts
rm -rf ${APPHOME}/web

sed -e "s;prefix =.*;prefix = ${PREFIX};g" \
	conf/app.ini > ${APPHOME}/conf/app.ini

sed -e "s;DEBUG;${LOG_LEVEL};g" \
	-e "s;HOSTNAME;${HOSTNAME};g" \
	conf/log.ini > ${APPHOME}/conf/log.ini

if ! [ -z "${LOG_SLACK_WEBHOOK}" ]; then
	LOG_WRITERS="stdout, file, access, dump, slack"
	sed -i \
		-e "s;LOG_SLACK_WEBHOOK;${LOG_SLACK_WEBHOOK};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi

cp -a ${APPNAME} ${APPHOME}/
cp -a tpls       ${APPHOME}/
cp -a txts       ${APPHOME}/
cp -a web        ${APPHOME}/
