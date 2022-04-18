package baremetal

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

/*
在这里将使用 从 secret 中获取的kubeadm命令 通过使用ssh sftp等方式去远程机器上执行

远程机器上需要做的事情
1. 安装 kubeadm kubelet kubectl
2. 安装 docker
   sftp 上传**.sh 脚本到服务器 在脚本中自动做一些事情（我只能简化这个脚本做的事情，细分的东西不能保证）
   运行 ./**.sh
3. 执行 kubeadm init / join 命令
4. 如果是高可用部署 则还要 判断该node是master还是CreateWorker
   如果是master还要将该node加入负载均衡器（maybe 我暂时也不知道这样是否合理）

*/

type BareMetal struct {
	cluster  string
	machine  string
	ipFamily clusterv1.ClusterIPFamily
}

// CreateControlPlaneNode 将裸金属创建为集群控制平面节点
func (metal *BareMetal) CreateControlPlaneNode() {

}

// CreateWorkerNode 将裸金属创建为集群工作节点
func (metal *BareMetal) CreateWorkerNode() {

}
