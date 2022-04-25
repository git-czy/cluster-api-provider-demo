kind delete cluster
kind create cluster
kubectl cluster-info --context kind-kind


kind load docker-image ccr.ccs.tencentyun.com/oldcc/cert-manager-webhook:v1.5.3
kind load docker-image ccr.ccs.tencentyun.com/oldcc/cert-manager-controller:v1.5.3
kind load docker-image ccr.ccs.tencentyun.com/oldcc/cert-manager-cainjector:v1.5.3



kind load docker-image ccr.ccs.tencentyun.com/oldcc/kubeadm-control-plane-controller:v1.1.3
kind load docker-image ccr.ccs.tencentyun.com/oldcc/kubeadm-bootstrap-controller:v1.1.3
kind load docker-image ccr.ccs.tencentyun.com/oldcc/cluster-api-controller:v1.1.3

kind load docker-image ccr.ccs.tencentyun.com/oldcc/cluster-api-provider-demo:latest
kind load docker-image ccr.ccs.tencentyun.com/oldcc/metal-node-controller:latest
kind load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0




cd /home/centos/go/src/cluster-api-metalnode || exit && make deploy
cd /home/centos/go/src/cluster-api-provider-demo || exit && make deploy



#kubectl apply -f /home/centos/go/src/cluster-api-metalnode/config/samples/metal_v1beta1_metalnode.yaml

#kubectl apply -f /home/centos/go/src/cluster-api-provider-demo/infrastructure-components.yaml

clusterctl init