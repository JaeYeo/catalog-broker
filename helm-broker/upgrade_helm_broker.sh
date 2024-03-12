helm upgrade helm-broker charts/helm-broker \
 --namespace helm-broker \
 --set global.containerRegistry.path="registry.systeer.com/" \
 --set global.helm_broker.dir="rancher/" \
 --set global.helm_broker.version=0.3 \
 --set global.helm_controller.dir="rancher/" \
 --set global.helm_controller.version=0.3

