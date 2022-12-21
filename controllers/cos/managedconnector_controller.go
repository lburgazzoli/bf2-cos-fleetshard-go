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
	errors2 "github.com/pkg/errors"
	camel2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/camel"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/conditions"
	meta2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/meta"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	camel "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/predicates"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
}

func NewManagedConnectorReconciler(mgr manager.Manager) (*ManagedConnectorReconciler, error) {
	r := &ManagedConnectorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		mgr:    mgr,
	}

	return r, r.SetupWithManager(mgr)
}

func (r *ManagedConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cos.ManagedConnector{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{},
			))).
		Owns(&corev1.Secret{}, builder.WithPredicates(
			predicate.Or(
				predicate.ResourceVersionChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{},
			))).
		Owns(&camel.KameletBinding{}, builder.WithPredicates(predicates.StatusChanged{})).
		Named("ManagedConnectorController").
		Complete(r)
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

	var connector cos.ManagedConnector
	var secret corev1.Secret

	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &connector); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name + "-deploy", Namespace: req.Namespace}, &secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// we'll ignore notification for resources not with the same UOW.
	// we'll need to wait for a new notification.

	if connector.Annotations[meta2.MetaUnitOfWork] != secret.Annotations[meta2.MetaUnitOfWork] {
		return ctrl.Result{
			RequeueAfter: 1 * time.Second,
		}, nil
	}

	//
	// Reconcile
	//

	rc := camel2.ReconciliationContext{
		C:      ctx,
		M:      r.mgr,
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Connector: connector.DeepCopy(),
		Secret:    secret.DeepCopy(),
	}

	if err := reconcileConnector(rc); err != nil {
		return ctrl.Result{}, err
	}

	//
	// Update connector
	//

	// TODO: must be properly computed or removed
	rc.Connector.Status.Phase = "Unknown"

	if err := controller.PatchStatus(ctx, r.Client, &connector, rc.Connector); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, err
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileConnector(rc camel2.ReconciliationContext) error {

	var binding camel.KameletBinding
	var bindingSecret corev1.Secret
	var bindingConfig corev1.ConfigMap

	if err := rc.GetDependant(&binding); err != nil {
		return errors2.Wrap(err, "failure loading dependant KameletBinding")
	}
	if err := rc.GetDependant(&bindingSecret); err != nil {
		return errors2.Wrap(err, "failure loading dependant KameletBinding secret")
	}
	if err := rc.GetDependant(&bindingConfig); err != nil {
		return errors2.Wrap(err, "failure loading dependant KameletBinding config")
	}

	if err := extractConditions(&rc.Connector.Status.Conditions, binding); err != nil {
		return errors2.Wrap(err, "unable to compute binding conditions")
	}

	meta.SetStatusCondition(&rc.Connector.Status.Conditions, readyCondition(*rc.Connector))

	//
	// Update binding & secret
	//

	switch rc.Connector.Spec.Deployment.DesiredState {
	case cos.DesiredStateReady:

		b, bs, bc, err := camel2.Reify(*rc.Connector, *rc.Secret)
		if err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(rc.Connector, &bs, rc.M.GetScheme()); err != nil {
			return errors2.Wrap(err, "unable to set binding secret controller reference")
		}
		if err := rc.PatchDependant(&bindingSecret, &bs); err != nil {
			return errors2.Wrap(err, "unable to patch binding secret")
		}

		if err := controllerutil.SetControllerReference(rc.Connector, &bc, rc.M.GetScheme()); err != nil {
			return errors2.Wrap(err, "unable to set binding config controller reference")
		}
		if err := rc.PatchDependant(&bindingConfig, &bc); err != nil {
			return errors2.Wrap(err, "unable to patch binding config")
		}

		if err := controllerutil.SetControllerReference(rc.Connector, &b, rc.M.GetScheme()); err != nil {
			return errors2.Wrap(err, "unable to set binding config controller reference")
		}
		if err := rc.PatchDependant(&binding, &b); err != nil {
			return errors2.Wrap(err, "unable to patch binding")
		}

		setReadyCondition(
			rc.Connector,
			metav1.ConditionTrue,
			conditions.ConditionReasonProvisioned,
			conditions.ConditionMessageProvisioned)

		rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
	case cos.DesiredStateStopped:
		setReadyCondition(
			rc.Connector,
			metav1.ConditionFalse,
			conditions.ConditionReasonStopping,
			conditions.ConditionMessageStopping)

		deleted := 0

		for _, r := range []client.Object{&binding, &bindingSecret, &bindingConfig} {
			if err := rc.DeleteDependant(r); err != nil {
				if !errors.IsNotFound(err) {
					deleted++
				}

				return err
			}
		}

		if deleted == 3 {
			setReadyCondition(
				rc.Connector,
				metav1.ConditionFalse,
				conditions.ConditionReasonStopped,
				conditions.ConditionMessageStopped)

			rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
		}
	case cos.DesiredStateDeleted:
		setReadyCondition(
			rc.Connector,
			metav1.ConditionFalse,
			conditions.ConditionReasonDeleting,
			conditions.ConditionMessageDeleting)

		deleted := 0

		for _, r := range []client.Object{&binding, &bindingSecret, &bindingConfig} {
			if err := rc.DeleteDependant(r); err != nil {
				if !errors.IsNotFound(err) {
					deleted++
				}

				return err
			}
		}

		if deleted == 3 {
			setReadyCondition(
				rc.Connector,
				metav1.ConditionFalse,
				conditions.ConditionReasonDeleted,
				conditions.ConditionMessageDeleted)

			rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
		}
	}

	return nil
}
