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
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	meta2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/predicates"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ManagedConnectorReconciler reconciles a ManagedConnector object
type ManagedConnectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	mgr    manager.Manager
	ctrl   controller.Controller
}

func NewManagedConnectorReconciler(mgr manager.Manager, ctrl controller.Controller) (*ManagedConnectorReconciler, error) {
	r := &ManagedConnectorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		mgr:    mgr,
		ctrl:   ctrl,
	}

	return r, r.Initialize(mgr)
}

func (r *ManagedConnectorReconciler) Initialize(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr).
		Named("ManagedConnectorController").
		For(&cos.ManagedConnector{}, builder.WithPredicates(
			predicate.Or(
				// TODO: add label selection
				// predicate.LabelSelectorPredicate(),
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{},
			))).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			&handler.EnqueueRequestForOwner{OwnerType: &cos.ManagedConnector{}},
			builder.WithPredicates(
				predicate.Or(
					// TODO: add label selection
					// predicate.LabelSelectorPredicate(),
					predicate.ResourceVersionChangedPredicate{},
					predicate.AnnotationChangedPredicate{},
					predicate.LabelChangedPredicate{},
				)))

	for i := range r.ctrl.Owned {
		c.Owns(
			r.ctrl.Owned[i],
			// TODO: add label selection
			// predicate.LabelSelectorPredicate(),
			builder.WithPredicates(predicates.StatusChanged{}))
	}

	return c.Complete(r)
}

//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=camel.apache.org,resources=kameletbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ManagedConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling", "namespace", req.Namespace, "name", req.Name)

	rc := controller.ReconciliationContext{
		C:      ctx,
		M:      r.mgr,
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Connector: &cos.ManagedConnector{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		},
		Secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		},
		ConfigMap: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		},
	}

	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, rc.Connector); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name + "-deploy", Namespace: req.Namespace}, rc.Secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name + "-deploy", Namespace: req.Namespace}, rc.ConfigMap); err != nil {
		if k8serrors.IsNotFound(err) {
			if err := r.Create(ctx, rc.ConfigMap); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// we'll ignore notification for resources not with the same UOW.
	// we'll need to wait for a new notification.

	if rc.Connector.Annotations[meta2.MetaUnitOfWork] != rc.Connector.Annotations[meta2.MetaUnitOfWork] {
		return ctrl.Result{
			RequeueAfter: 1 * time.Second,
		}, nil
	}

	//
	// Reconcile
	//

	// safe copy
	rc.Connector = rc.Connector.DeepCopy()

	if err := r.ctrl.ApplyFunc(rc); err != nil {
		return ctrl.Result{}, err
	}

	//
	// Update connector
	//

	// TODO: must be properly computed or removed
	rc.Connector.Status.Phase = "Unknown"

	if err := r.Status().Update(ctx, rc.Connector); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, err
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
