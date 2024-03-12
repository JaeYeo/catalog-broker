curl -X PUT http://cloudvmbroker.saas.sysmasterk8s.com/broker/v2/service_instances/vm01 \
	-H "Content-Type: application/json" \
	-d '{ "service_id": "55e94476-9d38-11ed-a8fc-0242ac120000", "plan_id": "55e94476-9d38-11ed-a8fc-0242ac120001", "parameters" : { "user_name" : "k8s", "tenant_name" : "k8s", "domain_name" : "admin_domain", "password" : "master77!!", "auth_url" : "http://172.40.1.195:5000/v3", "region" : "RegionOne", "key_pair" : "k8s_key", "network_uuid": "e0b8c6b6-45d6-4874-8500-2c35a59141d3", "security_groups" : "k8s_sec" } }' 
