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
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"metalnode/api/v1beta1"
	remote2 "metalnode/pkg/remote"
	util "metalnode/utils"
	"metalnode/utils/log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	INITIALIZING v1beta1.InitializationState = "INITIALIZING"
	CHECKING     v1beta1.InitializationState = "CHECKING"
	FAIL         v1beta1.InitializationState = "FAIL"
	SUCCESS      v1beta1.InitializationState = "SUCCESS"
	READY        bool                        = true
)

// MetalNodeReconciler reconciles a MetalNode object
type MetalNodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=metal.metal.node,resources=metalnodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MetalNode object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *MetalNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	metalNode := &v1beta1.MetalNode{}
	if err := r.Get(ctx, req.NamespacedName, metalNode); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	err := metalNode.Spec.NodeEndPoint.Validate()
	if err != nil {
		log.WithError(err)
		return ctrl.Result{}, err
	}

	// always update the status of the metal node,when leave reconcile
	defer func() {
		if metalNode.Status.InitializationState == SUCCESS {
			metalNode.Status.Ready = READY
		}
		if err := r.Status().Update(ctx, metalNode); err != nil {
			log.WithError(err)
		}
	}()

	// if metal node is INITIALIZING, do nothing
	if metalNode.Status.InitializationState == INITIALIZING {
		return ctrl.Result{}, nil
	}

	// if metal node is CHECKING, do nothing
	if metalNode.Status.InitializationState == CHECKING {
		return ctrl.Result{}, nil
	}

	if metalNode.Status.InitializationState == "" {
		metalNode.Status.InitializationState = INITIALIZING
		if err := r.Status().Update(ctx, metalNode); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.initMetal(ctx, metalNode); err != nil {
			return ctrl.Result{}, err
		}

		metalNode.Status.InitializationState = CHECKING
		if err := r.Status().Update(ctx, metalNode); err != nil {
			return ctrl.Result{}, err
		}
		// if metal node InitializationFailureReason is not empty, maybe means the initialization failed
		// so need to check the metal node is initialized or not(check docker kubelet kubeadm)
		if err := r.checkMetalNodeInitialized(ctx, metalNode); err != nil {
			metalNode.Status.InitializationState = FAIL
			if err := r.Status().Update(ctx, metalNode); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, errors.New("metal node initialization failed")
		}

		metalNode.Status.InitializationState = SUCCESS
		if err := r.Status().Update(ctx, metalNode); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetalNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.MetalNode{}).
		Complete(r)
}

// initMetal initializes the metal node
func (r *MetalNodeReconciler) initMetal(ctx context.Context, metalNode *v1beta1.MetalNode) error {
	host := []remote2.Host{
		{
			User:     metalNode.Spec.NodeEndPoint.SSHAuth.User,
			Password: metalNode.Spec.NodeEndPoint.SSHAuth.Password,
			Address:  metalNode.Spec.NodeEndPoint.Host,
			Port:     metalNode.Spec.NodeEndPoint.SSHAuth.Port,
			SSHKey:   metalNode.Spec.NodeEndPoint.SSHAuth.SSHKey,
		},
	}
	cmd := remote2.Cmd{
		Cmds: []string{
			"sudo chmod +x /tmp/init_k8s_env.sh",
			"sudo /tmp/init_k8s_env.sh",
		},
		FileUp: []remote2.File{
			{Src: "script/init_k8s_env.sh", Dst: "/tmp"},
		},
	}

	if metalNode.Spec.InitializationCmd != nil {
		cmd = *metalNode.Spec.InitializationCmd
	}

	errs := remote2.Run(host, cmd)
	if len(errs[metalNode.Spec.NodeEndPoint.Host]) != 0 {
		metalNode.Status.InitializationFailureReason = errs[metalNode.Spec.NodeEndPoint.Host]
		if err := r.Status().Update(ctx, metalNode); err != nil {
			return err
		}
	}
	return nil
}

// check metal node is already initialized
func (r *MetalNodeReconciler) checkMetalNodeInitialized(ctx context.Context, metalNode *v1beta1.MetalNode) error {
	host := []remote2.Host{
		{
			User:     metalNode.Spec.NodeEndPoint.SSHAuth.User,
			Password: metalNode.Spec.NodeEndPoint.SSHAuth.Password,
			Address:  metalNode.Spec.NodeEndPoint.Host,
			Port:     metalNode.Spec.NodeEndPoint.SSHAuth.Port,
			SSHKey:   metalNode.Spec.NodeEndPoint.SSHAuth.SSHKey,
		},
	}
	cmd := remote2.Cmd{
		Cmds: []string{
			"sudo docker version",
			"kubelet --version",
			"kubectl version",
		},
	}

	//if metalNode.Spec.InitializationCmd != nil {
	//	cmd = *metalNode.Spec.InitializationCmd
	//}

	errs := remote2.Run(host, cmd)

	// almost impossible to get one err when run kubectl version,but we don't care about it'
	checkErrs := util.SliceRemoveString(errs[metalNode.Spec.NodeEndPoint.Host], "The connection to the server localhost:8080 was refused - did you specify the right host or port?")

	if len(checkErrs) != 0 {
		metalNode.Status.CheckFailureReason = checkErrs
		if err := r.Status().Update(ctx, metalNode); err != nil {
			return err
		}
		return errors.New("metal node is initialized failed")
	}
	return nil
}
