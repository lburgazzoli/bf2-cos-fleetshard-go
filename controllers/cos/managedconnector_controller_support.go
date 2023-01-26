package cos

import (
	"context"
	"fmt"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/predicates"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func (r *ManagedConnectorReconciler) initialize(mgr ctrl.Manager) error {
	operatorTypeSelector, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: map[string]string{
			cosmeta.MetaOperatorType: r.options.Type,
		},
	})
	if err != nil {
		return errors.Wrapf(err, "unable to confiure operator-type label selector")
	}

	// TODO: refactor
	c := ctrl.NewControllerManagedBy(mgr).
		Named("ManagedConnectorController").
		For(&cos.ManagedConnector{}, builder.WithPredicates(
			predicate.And(
				operatorTypeSelector,
				predicate.Or(
					predicate.GenerationChangedPredicate{},
					predicate.AnnotationChangedPredicate{},
					predicate.LabelChangedPredicate{},
				)))).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			&handler.EnqueueRequestForOwner{OwnerType: &cos.ManagedConnector{}},
			builder.WithPredicates(
				predicate.And(
					operatorTypeSelector,
					predicate.Or(
						predicate.ResourceVersionChangedPredicate{},
						predicate.AnnotationChangedPredicate{},
						predicate.LabelChangedPredicate{},
					)))).
		Watches(
			&source.Kind{Type: &cos.ManagedConnectorOperator{}},
			handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
				mco, ok := a.(*cos.ManagedConnectorOperator)
				if !ok {
					r.l.Error(fmt.Errorf("type assertion failed: %v", a), "failed to retrieve ManagedConnectorOperator list")
					return nil
				}

				if mco.GetName() != r.options.ID {
					r.l.Info(
						"skip event",
						"operator-id", mco.GetName())

					return nil
				}

				requests, err := r.lookupManagedConnectorsForOperator()
				if err != nil {
					r.l.Error(err, "failed to retrieve ManagedConnectorOperator list")
					return nil
				}

				return requests
			}),
			builder.WithPredicates(
				predicate.And(
					operatorTypeSelector,
					predicate.GenerationChangedPredicate{}))).
		Owns(
			&camelv1alpha1.KameletBinding{},
			builder.WithPredicates(predicates.StatusChanged{}))

	return c.Complete(r)
}

func (r *ManagedConnectorReconciler) lookupManagedConnectorsForOperator() ([]reconcile.Request, error) {
	requests := make([]reconcile.Request, 0)

	list := &cos.ManagedConnectorList{}

	opts := []client.ListOption{
		client.MatchingLabels{
			cosmeta.MetaOperatorType: r.options.Type,
		},
		client.MatchingFields{
			"status.operatorId": r.options.ID,
		},
	}

	if err := r.List(context.Background(), list, opts...); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve ManagedConnectorOperator list")
	}

	for i := range list.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: resources.AsNamespacedName(&list.Items[i]),
		})
	}

	return requests, nil
}
