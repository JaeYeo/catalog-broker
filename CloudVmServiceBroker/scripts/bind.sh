curl -X PUT http://cloudvmbroker.saas.sysmasterk8s.com/broker/v2/service_instances/vm01/service_bindings/vm01_binding \
	-H "Content-Type: application/json" \
	-d '{ "service_id": "55e94476-9d38-11ed-a8fc-0242ac120000", "plan_id": "55e94476-9d38-11ed-a8fc-0242ac120001", "app_guid": "vm01" }'	
