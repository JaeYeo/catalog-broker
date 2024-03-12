helm install helm-broker charts/helm-broker \
 --namespace helm-broker --create-namespace \
 --set global.containerRegistry.path="registry.systeer.com/" \
 --set global.helm_broker.dir="rancher/" \
 --set global.helm_broker.version=latest5 \
 --set global.helm_controller.dir="rancher/" \
 --set global.helm_controller.version=latest5

