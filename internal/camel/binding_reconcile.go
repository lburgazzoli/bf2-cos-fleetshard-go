package camel

import (
	kamelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/conditions"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Apply(rc controller.ReconciliationContext) error {

	binding := kamelv1alpha1.KameletBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Connector.Name,
			Namespace: rc.Connector.Namespace,
		},
	}
	bindingSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Connector.Name + "-secret",
			Namespace: rc.Connector.Namespace,
		},
	}
	bindingConfig := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Connector.Name + "-config",
			Namespace: rc.Connector.Namespace,
		},
	}

	if err := rc.GetDependant(&binding); err != nil {
		return errors.Wrap(err, "failure loading dependant KameletBinding")
	}
	if err := rc.GetDependant(&bindingSecret); err != nil {
		return errors.Wrap(err, "failure loading dependant KameletBinding secret")
	}
	if err := rc.GetDependant(&bindingConfig); err != nil {
		return errors.Wrap(err, "failure loading dependant KameletBinding config")
	}

	//
	// Update connector
	//

	if err := extractConditions(&rc.Connector.Status.Conditions, binding); err != nil {
		return errors.Wrap(err, "unable to compute binding conditions")
	}

	conditions.Set(&rc.Connector.Status.Conditions, conditions.Ready(*rc.Connector))

	//
	// Update binding & secret
	//

	switch rc.Connector.Spec.Deployment.DesiredState {
	case cos.DesiredStateReady:

		b, bs, bc, err := reify(&rc)
		if err != nil {
			return err
		}

		if err := patchDependant(rc, &bindingSecret, &bs); err != nil {
			return errors.Wrap(err, "unable to reconcile binding secrete")
		}
		if err := patchDependant(rc, &bindingConfig, &bc); err != nil {
			return errors.Wrap(err, "unable to reconcile binding config")
		}
		if err := patchDependant(rc, &binding, &b); err != nil {
			return errors.Wrap(err, "unable to reconcile binding")
		}

		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = conditions.ConditionReasonProvisioned
			condition.Message = conditions.ConditionMessageProvisioned
			condition.ResourceRevision = rc.Connector.Spec.Deployment.DeploymentResourceVersion
		})

		rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
	case cos.DesiredStateStopped:
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonStopping
			condition.Message = conditions.ConditionMessageStopping
			condition.ResourceRevision = rc.Connector.Spec.Deployment.DeploymentResourceVersion
		})

		deleted := 0

		for _, r := range []client.Object{&binding, &bindingSecret, &bindingConfig} {
			if err := rc.DeleteDependant(r); err != nil {
				if !k8serrors.IsNotFound(err) {
					return err
				}

				deleted++
			}
		}

		if deleted == 3 {
			conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
				condition.Status = metav1.ConditionFalse
				condition.Reason = conditions.ConditionReasonStopped
				condition.Message = conditions.ConditionMessageStopped
				condition.ResourceRevision = rc.Connector.Spec.Deployment.DeploymentResourceVersion
			})

			rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
		}
	case cos.DesiredStateDeleted:
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonDeleting
			condition.Message = conditions.ConditionMessageDeleting
			condition.ResourceRevision = rc.Connector.Spec.Deployment.DeploymentResourceVersion
		})

		deleted := 0

		for _, r := range []client.Object{&binding, &bindingSecret, &bindingConfig} {
			if err := rc.DeleteDependant(r); err != nil {
				if !k8serrors.IsNotFound(err) {
					return err
				}

				deleted++
			}
		}

		if deleted == 3 {
			conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
				condition.Status = metav1.ConditionFalse
				condition.Reason = conditions.ConditionReasonDeleted
				condition.Message = conditions.ConditionMessageDeleted
				condition.ResourceRevision = rc.Connector.Spec.Deployment.DeploymentResourceVersion
			})

			rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
		}
	}

	for i := range rc.Connector.Status.Conditions {
		rc.Connector.Status.Conditions[i].ObservedGeneration = rc.Connector.Status.ObservedGeneration
	}

	return nil
}
