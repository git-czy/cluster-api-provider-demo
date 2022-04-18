package constants

// role value constants
const (
	ControlPlaneNodeRoleValue = "control-plane"
	WorkerNodeRoleValue       = "worker"
	EtcdRoleValue             = "etcd"
	LoadBalancerRoleValue     = "load-balancer"
)

//  condition type constants
const (
	// ControlPlaneEndPointSetCondition ConditionTypeNodeReady is set when the control plane endpoints are set
	ControlPlaneEndPointSetCondition = "ControlPlaneEndPointSet"

	// MetalNodeReadyCondition denotes the metal node is ready to be used
	MetalNodeReadyCondition = "MetalNodeReady"
)

// condition reason constants
const (
	// NoMetalNodeFoundReason is set when no metal nodes are not found
	NoMetalNodeFoundReason = "NoMetalNodeFound"
)
