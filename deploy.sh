#!/bin/bash -ex

APPNAME=xdemo
APPHOME=/app/${APPNAME}
PREFIX=
LOG_WRITERS="stdout, file, dump, access"

#LOG_OPENSEARCH_BATCH_WEBHOOK=https://localhost:9200/xdemo_logs/_bulk
#LOG_OPENSEARCH_ACCESS_WEBHOOK=https://localhost:9200/xdemo_access/_bulk

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
	LOG_WRITERS="${LOG_WRITERS}, slack"
	sed -i \
		-e "s;LOG_SLACK_WEBHOOK;${LOG_SLACK_WEBHOOK};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi
if ! [ -z "${LOG_OPENSEARCH_BATCH_WEBHOOK}" ]; then
	LOG_WRITERS="${LOG_WRITERS}, opensearch"
	sed -i \
		-e "s;LOG_OPENSEARCH_BATCH_WEBHOOK;${LOG_OPENSEARCH_BATCH_WEBHOOK};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi
if ! [ -z "${LOG_OPENSEARCH_ACCESS_WEBHOOK}" ]; then
	LOG_WRITERS="${LOG_WRITERS}, accessos"
	sed -i \
		-e "s;LOG_OPENSEARCH_ACCESS_WEBHOOK;${LOG_OPENSEARCH_ACCESS_WEBHOOK};g" \
		-e "s;writer =.*;writer = ${LOG_WRITERS};g" \
		${APPHOME}/conf/log.ini
fi

cp -a ${APPNAME} ${APPHOME}/
cp -a tpls       ${APPHOME}/
cp -a txts       ${APPHOME}/
cp -a web        ${APPHOME}/
