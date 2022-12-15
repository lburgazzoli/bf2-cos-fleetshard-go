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

package cos

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
)

// ManagedConnectorReconciler reconciles a ManagedConnector object
type ManagedConnectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	mgr    manager.Manager
}

func NewManagedConnectorReconciler(mgr manager.Manager) (*ManagedConnectorReconciler, error) {
	r := &ManagedConnectorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		mgr:    mgr,
	}

	return r, r.SetupWithManager(mgr)
}

//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors/finalizers,verbs=update

func (r *ManagedConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosv2.ManagedConnector{}).
		Complete(r)
}
