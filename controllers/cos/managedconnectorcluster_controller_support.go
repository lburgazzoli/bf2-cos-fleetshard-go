package cos

import (
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *ManagedConnectorClusterReconciler) initialize(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr).
		Named("ManagedConnectorClusterController").
		For(&cos.ManagedConnectorCluster{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{},
			)))

	return c.Complete(r)
}
