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
	"sigs.k8s.io/cluster-api/util/conditions"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrav1 "cluster-api-provider-demo/api/v1beta1"
	"cluster-api-provider-demo/constants"
	metav1beta1 "cluster-api-provider-demo/metalnode/api/v1beta1"
	"cluster-api-provider-demo/utils/log"
)

// DemoClusterReconciler reconciles a DemoCluster object
type DemoClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=democlusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=democlusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=democlusters/finalizers,verbs=get;update;patch
//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes,verbs=get;list
//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DemoCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DemoClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {

	// todo 1 Fetch the DemoCluster instance.
	demoCluster := &infrav1.DemoCluster{}
	if err := r.Get(ctx, req.NamespacedName, demoCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// todo 2 Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, demoCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, demoCluster) {
		log.Info("demoCluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	// todo 3 Initialize the patch helper
	patchHelper, err := patch.NewHelper(demoCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the DemoCluster object and status after each reconciliation.
	defer func() {
		if err := patchDemoCluster(ctx, patchHelper, demoCluster); err != nil {
			log.Error("failed to patch demoMachine", err)
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// todo 4 Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(demoCluster, infrav1.ClusterFinalizer) {
		controllerutil.AddFinalizer(demoCluster, infrav1.ClusterFinalizer)
		return ctrl.Result{}, nil
	}

	// todo 5 Handle deleted clusters
	if !demoCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, demoCluster)
	}

	// todo 6 Handle non-deleted clusters
	return r.reconcileNormal(ctx, demoCluster, cluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DemoClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DemoCluster{}).
		Complete(r)
}

// reconcileDelete reconcile demoCluster delete
func (r *DemoClusterReconciler) reconcileDelete(ctx context.Context, demoCluster *infrav1.DemoCluster) (ctrl.Result, error) {
	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(demoCluster, infrav1.ClusterFinalizer)

	return ctrl.Result{}, nil
}

// reconcileNormal reconcile demoCluster normal
func (r *DemoClusterReconciler) reconcileNormal(ctx context.Context, demoCluster *infrav1.DemoCluster, cluster *clusterv1.Cluster) (ctrl.Result, error) {
	// todo reconcile normal logic here

	log.Info("preparing load balancer")
	// todo 这里一般回来检查 demoCluster是否配置好,按照docker provider的做法 应该创建一个load_balancer
	// todo 事实上从官方搭建高可用文档的案列来看 也是这么做的
	// todo 如果不是高可用部署，那么无需安装负载均衡器，还未找到查看是否是高可用部署的方法，搁置

	// todo 应该在此处为demoCluster设置ControlPlaneEndpoint，在docker-provider中设置的是负载均衡器的 地址
	// todo 如果不使用 docker 容器安装负载均衡器或者不使用高可用安装集群，那么应该直接使用裸机的IP，port为6443

	// 目前先不考虑高可用部署，从所有metalnode中选择一个作为controlplane使用其 ip：6443设置为controlplane endpoint
	metalNodeList := &metav1beta1.MetalNodeList{}
	if err := r.Client.List(ctx, metalNodeList, client.InNamespace(demoCluster.Namespace)); err != nil {
		conditions.MarkFalse(demoCluster, constants.ControlPlaneEndPointSetCondition, constants.NoMetalNodeFoundReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	if len(metalNodeList.Items) == 0 {
		conditions.MarkFalse(demoCluster, constants.ControlPlaneEndPointSetCondition, constants.NoMetalNodeFoundReason, clusterv1.ConditionSeverityWarning, "no metal node found")
		return ctrl.Result{}, fmt.Errorf("no metalnode found")
	}

	controlPlaneNode := &metalNodeList.Items[0]

	// set demoCluster controlPlaneEndpoint
	demoCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
		Host: controlPlaneNode.Spec.NodeEndPoint.Host,
		Port: 6443,
	}

	// set controlPlaneNode Role and RefCluster
	controlPlaneNode.SetRole(constants.ControlPlaneNodeRoleValue)
	controlPlaneNode.Status.RefCluster = cluster.Name
	if err := r.Status().Update(ctx, controlPlaneNode); err != nil {
		return ctrl.Result{}, err
	}

	// Mark the demoCluster ready
	demoCluster.Status.Ready = true
	conditions.MarkTrue(demoCluster, constants.ControlPlaneEndPointSetCondition)

	return ctrl.Result{}, nil
}

// patchDemoCluster will patch the DemoCluster
func patchDemoCluster(ctx context.Context, patchHelper *patch.Helper, demoCluster *infrav1.DemoCluster) error {
	return patchHelper.Patch(
		ctx,
		demoCluster,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			constants.ControlPlaneEndPointSetCondition,
		}})
}
