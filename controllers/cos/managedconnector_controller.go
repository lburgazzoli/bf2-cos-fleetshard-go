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
	"fmt"
	errors2 "github.com/pkg/errors"
	camel2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/camel"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
				predicates.AnnotationChanged{
					Name: "cos.bf2.dev/uow",
				},
			))).
		Owns(&corev1.Secret{}, builder.WithPredicates(
			predicate.Or(
				predicate.ResourceVersionChangedPredicate{},
				predicates.AnnotationChanged{
					Name: "cos.bf2.dev/uow",
				},
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

func (r *ManagedConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	named := types.NamespacedName{Name: req.Name, Namespace: req.Namespace}
	connector := cos.ManagedConnector{}
	secret := corev1.Secret{}

	if err := r.Get(ctx, named, &connector); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.Get(ctx, named, &secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// we'll ignore notification for resources not with the same UOW.
	// we'll need to wait for a new notification.

	if connector.Annotations["cos.bf2.dev/uow"] != secret.Annotations["cos.bf2.dev/uow"] {
		return ctrl.Result{
			RequeueAfter: 1 * time.Second,
		}, nil
	}

	binding := camel.KameletBinding{}
	bindingSecret := corev1.Secret{}
	bindingConfig := corev1.ConfigMap{}

	if err := r.Get(ctx, named, &binding); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}
	if err := r.Get(ctx, named, &bindingSecret); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}
	if err := r.Get(ctx, named, &bindingConfig); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	c := connector.DeepCopy()

	if err := extractConditions(&connector.Status.Conditions, binding); err != nil {
		return ctrl.Result{}, errors2.Wrap(err, "unable to compute binding conditions")
	}

	meta.SetStatusCondition(&c.Status.Conditions, readyCondition(connector))

	//
	// Update binding & secret
	//

	switch connector.Spec.Deployment.DesiredState {
	case cos.DesiredStateReady:

		b, bs, bc, err := camel2.Reify(connector, secret)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.PatchSubresource(ctx, c, &bindingSecret, &bs); err != nil {
			return ctrl.Result{}, errors2.Wrap(err, "unable to patch binding secret")
		}
		if err := r.PatchSubresource(ctx, c, &bindingConfig, &bc); err != nil {
			return ctrl.Result{}, errors2.Wrap(err, "unable to patch binding config")
		}
		if err := r.PatchSubresource(ctx, c, &binding, &b); err != nil {
			return ctrl.Result{}, errors2.Wrap(err, "unable to patch binding")
		}

		controller.UpdateStatusCondition(&connector.Status.Conditions, "Ready", func(condition *metav1.Condition) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = "Provisioned"
			condition.Message = "Provisioned"
		})

		c.Status.Deployment = c.Spec.Deployment
		c.Status.ObservedGeneration = c.Generation
	case cos.DesiredStateStopped:
		controller.UpdateStatusCondition(&connector.Status.Conditions, "Ready", func(condition *metav1.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "Stopping"
			condition.Message = "Stopping"
		})

		if err := r.Delete(ctx, &binding); err != nil {
			if errors.IsNotFound(err) {
				controller.UpdateStatusCondition(&connector.Status.Conditions, "Ready", func(condition *metav1.Condition) {
					condition.Status = metav1.ConditionFalse
					condition.Reason = "Stopped"
					condition.Message = "Stopped"
				})

				c.Status.Deployment = c.Spec.Deployment
				c.Status.ObservedGeneration = c.Generation
			} else {
				return ctrl.Result{}, err
			}
		}
	case cos.DesiredStateDeleted:
		controller.UpdateStatusCondition(&connector.Status.Conditions, "Ready", func(condition *metav1.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "Deleting"
			condition.Message = "Deleting"
		})

		if err := r.Delete(ctx, &binding); err != nil {

			if errors.IsNotFound(err) {
				controller.UpdateStatusCondition(&connector.Status.Conditions, "Ready", func(condition *metav1.Condition) {
					condition.Status = metav1.ConditionFalse
					condition.Reason = "Deleted"
					condition.Message = "Deleted"
				})

				c.Status.Deployment = c.Spec.Deployment
				c.Status.ObservedGeneration = c.Generation
			} else {
				return ctrl.Result{}, err
			}
		}
	}

	//
	// Update connector
	//

	if err := controller.PatchStatus(ctx, r.Client, &connector, c); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ManagedConnectorReconciler) PatchSubresource(
	ctx context.Context,
	connector *cos.ManagedConnector,
	source client.Object,
	target client.Object,
) error {

	if err := controllerutil.SetControllerReference(connector, target, r.Scheme); err != nil {
		return err
	}

	target.GetAnnotations()["cos.bf2.dev/deployment.revision"] = fmt.Sprintf("%d", connector.Spec.Deployment.DeploymentResourceVersion)

	return controller.Patch(ctx, r.Client, source, target)
}
