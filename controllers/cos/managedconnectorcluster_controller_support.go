package cos

import (
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *ManagedConnectorClusterReconciler) initialize(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr).
		Named("ManagedConnectorClusterController").
		For(&cos.ManagedConnector{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{},
			)))

	for i := range r.options.Reconciler.Owned {
		c.Owns(
			r.options.Reconciler.Owned[i],
			// TODO: add label selection
			// predicate.LabelSelectorPredicate(),
			builder.WithPredicates(predicates.StatusChanged{}))
	}

	return c.Complete(r)
}
