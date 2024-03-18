#/bin/bash

day=`date +%d`
message=`tail -n 20 /root/egovp-rabbitmq-k8s/logs/logback-2021-03-${day}.0.log | grep "of durable queue 'ks.queue' in vhost '/' is down or inaccessible"`

echo ${day}
echo ${message}

if [ "${message}" != "" ]
then
	echo "서비스 재시작"
	/root/egovp-rabbitmq-k8s/stopK8sMgmtService.sh
	/root/egovp-rabbitmq-k8s/startK8sMgmtService.sh
else
	echo "서비스 정상"
fi
