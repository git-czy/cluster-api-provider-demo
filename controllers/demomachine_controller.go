/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/git-czy/cluster-api-provider-demo/constants"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1beta1 "github.com/git-czy/cluster-api-metalnode/api/v1beta1"
	infrav1 "github.com/git-czy/cluster-api-provider-demo/api/v1beta1"
	"github.com/git-czy/cluster-api-provider-demo/utils/log"
)

// DemoMachineReconciler reconciles a DemoMachine object
type DemoMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=demomachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=demomachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=demomachines/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;machines,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch
//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes,verbs=get;list
//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes/status,verbs=get;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DemoMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DemoMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {

	// todo 1 Fetch the DemoMachine instance.
	demoMachine := &infrav1.DemoMachine{}
	if err := r.Client.Get(ctx, req.NamespacedName, demoMachine); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// todo 2 Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, demoMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machine == nil {
		log.Info("Waiting for Machine Controller to set OwnerRef on DemoMachine")
		return ctrl.Result{}, nil
	}

	l := log.With("machine", machine.Name)
	// todo 3 Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		l.Info("DockerMachine owner Machine is missing cluster label or cluster does not exist")
		return ctrl.Result{}, err
	}
	if cluster == nil {
		l.Info(fmt.Sprintf("Please associate this machine with a cluster using the label %s: <name of cluster>", clusterv1.ClusterLabelName))
		return ctrl.Result{}, nil
	}

	l = l.With("cluster", cluster.Name)

	// todo 4 Return early if the object or Cluster is paused.
	if annotations.IsPaused(cluster, demoMachine) {
		l.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// todo 5 Fetch the Demo Cluster.
	demoCluster := &infrav1.DemoCluster{}
	demoClusterName := client.ObjectKey{
		Namespace: demoMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, demoClusterName, demoCluster); err != nil {
		l.Info("DemoCluster is not available yet")
		return ctrl.Result{}, nil
	}

	l = l.With("demo-cluster", demoCluster.Name)

	// todo 6 Initialize the patch helper
	patchHelper, err := patch.NewHelper(demoMachine, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the demoMachine object and status after each reconciliation.
	defer func() {
		if err := patchDemoMachine(ctx, patchHelper, demoMachine); err != nil {
			log.Errorf("failed to patch demoMachine")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// todo 7 Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(demoMachine, infrav1.MachineFinalizer) {
		controllerutil.AddFinalizer(demoMachine, infrav1.MachineFinalizer)
		return ctrl.Result{}, nil
	}

	// todo 8 Check if the infrastructure is ready, otherwise return and wait for the cluster object to be updated
	if !cluster.Status.InfrastructureReady {
		log.Info("Waiting for DemoCluster Controller to create cluster infrastructure")
		return ctrl.Result{}, nil
	}

	// todo 9 Handle deleted machines
	if !demoMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machine, demoMachine, demoCluster)
	}

	return r.reconcileNormal(ctx, machine, cluster, demoMachine, demoCluster, l)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DemoMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DemoMachine{}).
		Complete(r)
}

// reconcileDelete reconcile demoMachine delete
func (r *DemoMachineReconciler) reconcileDelete(ctx context.Context, machine *clusterv1.Machine, demoMachine *infrav1.DemoMachine, demoCluster *infrav1.DemoCluster) (ctrl.Result, error) {

	patchHelper, err := patch.NewHelper(demoMachine, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	conditions.MarkFalse(demoMachine, constants.MetalNodeReadyCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	if err := patchDemoMachine(ctx, patchHelper, demoMachine); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to patch demoMachine")
	}

	metalNode := &metav1beta1.MetalNode{}
	metalNodeList, err := r.getMetalNodes(ctx, demoCluster)
	if err != nil {
		conditions.MarkFalse(demoCluster, constants.MetalNodeReadyCondition, constants.DeletingReason, clusterv1.ConditionSeverityInfo, err.Error())
		return ctrl.Result{}, err
	}

	for _, node := range metalNodeList.Items {
		if string(node.UID) == demoMachine.Spec.ProviderID {
			metalNode = &node
			break
		}
	}

	// todo operate control plane machine

	// reset metalNode
	metalNode.ResetMetalNode()
	if err := r.Client.Status().Update(ctx, metalNode); err != nil {
		conditions.MarkFalse(demoCluster, constants.MetalNodeReadyCondition, constants.DeletingReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	controllerutil.RemoveFinalizer(demoMachine, infrav1.MachineFinalizer)
	return ctrl.Result{}, nil
}

// reconcileNormal reconcile demoMachine normal
func (r *DemoMachineReconciler) reconcileNormal(ctx context.Context, machine *clusterv1.Machine, cluster *clusterv1.Cluster, demoMachine *infrav1.DemoMachine, demoCluster *infrav1.DemoCluster, l log.Logger) (ctrl.Result, error) {
	metalNode := &metav1beta1.MetalNode{}
	metalNodeList, err := r.getMetalNodes(ctx, demoCluster)
	if err != nil {
		conditions.MarkFalse(demoCluster, constants.MetalNodeReadyCondition, constants.NoMetalNodeFoundReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	// if the machine is already provisioned, return
	if demoMachine.Spec.ProviderID != "" {
		// ensure ready state is set.
		// This is required after move, because status is not moved to the target cluster.
		demoMachine.Status.Ready = true
		for _, node := range metalNodeList.Items {
			if string(node.UID) == demoMachine.Spec.ProviderID {
				conditions.MarkTrue(demoMachine, constants.MetalNodeReadyCondition)
				metalNode = &node
				break
			}
		}
		setMachineAddress(demoMachine, metalNode)
		return ctrl.Result{}, nil
	}

	// Make sure bootstrap data is available and populated.
	if machine.Spec.Bootstrap.DataSecretName == nil {
		if !util.IsControlPlaneMachine(machine) && !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
			l.Info("Waiting for the control plane to be initialized")
			conditions.MarkFalse(demoMachine, constants.MetalNodeReadyCondition, clusterv1.WaitingForControlPlaneAvailableReason, clusterv1.ConditionSeverityInfo, "")
			return ctrl.Result{}, nil
		}

		l.Info("Waiting for the Bootstrap provider controller to set bootstrap data")
		conditions.MarkFalse(demoMachine, constants.MetalNodeReadyCondition, constants.WaitingForBootstrapDataReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// get metalNode hosting the machine
	role := constants.WorkerNodeRoleValue
	if util.IsControlPlaneMachine(machine) {
		role = constants.ControlPlaneNodeRoleValue
	}

	for _, node := range metalNodeList.Items {
		if node.ContainRole(role) && node.GetRefCluster() == cluster.Name && node.IsReady() {
			metalNode = &node
			break
		}
	}
	// if no metalNode found, return
	if metalNode == nil {
		conditions.MarkFalse(demoMachine, constants.MetalNodeReadyCondition, constants.NoMetalNodeFoundReason, clusterv1.ConditionSeverityWarning, "no metal node found")
		return ctrl.Result{}, fmt.Errorf("no metalnode found")
	}

	// if the machine isn't bootstrapped, only update the metalNode status(bootstrapData) to note metalNode run bootstrap data
	if !demoMachine.Status.Bootstrapped {
		if machine.Spec.Bootstrap.DataSecretName == nil {
			conditions.MarkFalse(demoMachine, constants.BootstrapDataAvailableCondition, constants.BootstrapDataNotAvailableReason, clusterv1.ConditionSeverityInfo, "bootstrap data not available")
			return ctrl.Result{}, errors.New("error retrieving bootstrap data: linked Machine's bootstrap.dataSecretName is nil")
		}
		conditions.MarkTrue(demoMachine, constants.BootstrapDataAvailableCondition)

		// update metalNode status
		metalNode.Status.DataSecretName = *machine.Spec.Bootstrap.DataSecretName
		err := r.Client.Status().Update(ctx, metalNode)
		if err != nil {
			l.Errorln("failed to update metalNode status", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Update the BootstrapSucceededCondition condition
	conditions.MarkTrue(demoMachine, constants.BootstrapSucceededCondition)
	setMachineAddress(demoMachine, metalNode)

	// Set ProviderID so the Cluster API Machine Controller can pull it
	demoMachine.Spec.ProviderID = string(metalNode.UID)
	demoMachine.Status.Ready = true
	conditions.MarkTrue(demoMachine, constants.MetalNodeReadyCondition)
	return ctrl.Result{}, nil
}

// setMachineAddress gets the address from the metal node .spec.NodeEndPoint.Host and sets it on the Machine object.
func setMachineAddress(demoMachine *infrav1.DemoMachine, metalNode *metav1beta1.MetalNode) {
	demoMachine.Status.Addresses = []clusterv1.MachineAddress{
		{
			Type:    clusterv1.MachineHostName,
			Address: metalNode.Name,
		},
		{
			Type:    clusterv1.MachineExternalIP,
			Address: metalNode.Spec.NodeEndPoint.Host,
		},
	}
}

// patchDemoCluster will patch the DemoCluster
func patchDemoMachine(ctx context.Context, patchHelper *patch.Helper, demoMachine *infrav1.DemoMachine) error {
	return patchHelper.Patch(ctx, demoMachine)
}

// getMetalNodes returns a list of metal nodes
func (r *DemoMachineReconciler) getMetalNodes(ctx context.Context, demoCluster *infrav1.DemoCluster) (*metav1beta1.MetalNodeList, error) {
	metalNodes := &metav1beta1.MetalNodeList{}
	if err := r.Client.List(ctx, metalNodes, client.InNamespace(demoCluster.Namespace)); err != nil {
		return nil, err
	}
	if len(metalNodes.Items) == 0 {
		return nil, fmt.Errorf("no metalnode found")
	}
	return metalNodes, nil
}
