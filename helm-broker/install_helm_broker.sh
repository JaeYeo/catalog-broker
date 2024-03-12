helm install helm-broker charts/helm-broker \
 --namespace helm-broker --create-namespace \
 --set global.containerRegistry.path="registry.sysmasterk8s-v3.com/" \
 --set global.helm_broker.dir="rancher/" \
 --set global.helm_broker.version=0.8 \
 --set global.helm_controller.dir="rancher/" \
 --set global.helm_controller.version=0.8 \
 --set webhook.image="registry.sysmasterk8s-v3.com/helm-broker/helm-broker-webhook:0.8" \
 --set etcd-stateful.etcd.image="registry.sysmasterk8s-v3.com/helm-broker/etcd" \
 --set etcd-stateful.etcd.imageTag="v3.3.9" \
 --set etcd-stateful.tlsSetup.image="registry.sysmasterk8s-v3.com/helm-broker/etcd-tls-setup" \
 --set etcd-stateful.tlsSetup.imageTag="0.3.367" \
 --set imageRegistry="registry.sysmasterk8s-v3.com"

