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

	// BootstrapDataAvailableCondition denotes the bootstrap data is available
	BootstrapDataAvailableCondition = "BootstrapDataAvailable"

	// BootstrapSucceededCondition denotes the bootstrap succeeded
	BootstrapSucceededCondition = "BootstrapSucceeded"
)

// condition reason constants
const (
	//NoMetalNodeFoundReason (Severity=Warning) is set when no metal nodes are not found
	NoMetalNodeFoundReason = "NoMetalNodeFound"

	// WaitingForBootstrapDataReason (Severity=Info) documents a DemoMachine waiting for the bootstrap
	// script to be ready before starting to create the container that provides the DockerMachine infrastructure.
	WaitingForBootstrapDataReason = "WaitingForBootstrapData"

	// BootstrapDataNotAvailableReason (Severity=Info) documents a DemoMachine waiting for the control plane
	BootstrapDataNotAvailableReason = "BootstrapDataNotAvailable"

	// DeletingReason (Severity=Info) documents a condition not in Status=True because the underlying object it is currently being deleted.
	DeletingReason = "Deleting"

	//WaitingForMetalNodeReadyReason (Severity=Info) documents a DemoMachine waiting for the metal node initialized
	WaitingForMetalNodeReadyReason = "WaitingForMetalNodeReady"

	//WaitingForMetalNodeBootstrapReason (Severity=Info) documents a DemoMachine waiting for the metal node bootstrap
	WaitingForMetalNodeBootstrapReason = "WaitingForMetalNodeBootstrap"
)
