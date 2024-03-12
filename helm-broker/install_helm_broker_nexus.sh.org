helm install helm-broker https://k8s-nexus.spaasta.com/repository/helm-hosted/helm-broker-0.1.0.tgz \
 --namespace helm-broker --create-namespace \
 --set global.containerRegistry.path="registry.systeer.com/" \
 --set global.helm_broker.dir="rancher/" \
 --set global.helm_broker.version=0.3 \
 --set global.helm_controller.dir="rancher/" \
 --set global.helm_controller.version=0.3

