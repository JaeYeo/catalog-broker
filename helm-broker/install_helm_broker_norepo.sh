helm install helm-broker charts/helm-broker \
 --namespace helm-broker --create-namespace \
 --set global.containerRegistry.path="registry.systeer.com/" \
 --set global.helm_broker.dir="rancher/" \
 --set global.helm_broker.version=0.7 \
 --set global.helm_controller.dir="rancher/" \
 --set global.helm_controller.version=0.7 \
 --set webhook.image="registry.systeer.com/rancher/helm-broker-webhook:0.3" \
 --set etcd-stateful.etcd.image="registry.systeer.com/rancher/etcd" \
 --set etcd-stateful.etcd.imageTag="v3.3.9" \
 --set etcd-stateful.tlsSetup.image="registry.systeer.com/rancher/etcd-tls-setup" \
 --set etcd-stateful.tlsSetup.imageTag="0.3.367"

