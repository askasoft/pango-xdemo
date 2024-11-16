#!/bin/bash -ex

APPNAME=xdemo
APPHOME=/app/${APPNAME}
PREFIX=
LOG_WRITERS="stdout, textfile, jsonfile, httpdump, xalfile, xajfile"

#LOG_OPENSEARCH_APPLOG=https://localhost:9200/xdemo_applog/_bulk
#LOG_OPENSEARCH_ACCESS=https://localhost:9200/xdemo_access/_bulk

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
	conf/log.ini > ${APPHOME}/conf/log.ini

cp -a conf/config.csv ${APPHOME}/conf/
cp -a conf/schema.sql ${APPHOME}/conf/

if ! [ -z "${LOG_SLACK_WEBHOOK}" ]; then
	LOG_WRITERS="${LOG_WRITERS}, slack"
	sed -i \
		-e "s;LOG_SLACK_WEBHOOK;${LOG_SLACK_WEBHOOK};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi
if ! [ -z "${LOG_OPENSEARCH_APPLOG}" ]; then
	LOG_WRITERS="${LOG_WRITERS}, appos"
	sed -i \
		-e "s;LOG_OPENSEARCH_APPLOG;${LOG_OPENSEARCH_APPLOG};g" \
		-e "s;LOG_OPENSEARCH_USERNAME;${LOG_OPENSEARCH_USERNAME};g" \
		-e "s;LOG_OPENSEARCH_PASSWORD;${LOG_OPENSEARCH_PASSWORD};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi
if ! [ -z "${LOG_OPENSEARCH_ACCESS}" ]; then
	LOG_WRITERS="${LOG_WRITERS}, xajos"
	sed -i \
		-e "s;LOG_OPENSEARCH_ACCESS;${LOG_OPENSEARCH_ACCESS};g" \
		-e "s;LOG_OPENSEARCH_USERNAME;${LOG_OPENSEARCH_USERNAME};g" \
		-e "s;LOG_OPENSEARCH_PASSWORD;${LOG_OPENSEARCH_PASSWORD};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi

cp -a ${APPNAME}* ${APPHOME}/
#cp -a tpls        ${APPHOME}/
#cp -a txts        ${APPHOME}/
#cp -a web         ${APPHOME}/
