package camel

import (
	kamelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/conditions"

	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Reconcile(rc controller.ReconciliationContext) error {

	var binding kamelv1alpha1.KameletBinding
	var bindingSecret corev1.Secret
	var bindingConfig corev1.ConfigMap

	if err := rc.GetDependant(&binding); err != nil {
		return errors.Wrap(err, "failure loading dependant KameletBinding")
	}
	if err := rc.GetDependant(&bindingSecret); err != nil {
		return errors.Wrap(err, "failure loading dependant KameletBinding secret")
	}
	if err := rc.GetDependant(&bindingConfig); err != nil {
		return errors.Wrap(err, "failure loading dependant KameletBinding config")
	}

	if err := extractConditions(&rc.Connector.Status.Conditions, binding); err != nil {
		return errors.Wrap(err, "unable to compute binding conditions")
	}

	meta.SetStatusCondition(&rc.Connector.Status.Conditions, conditions.Ready(*rc.Connector))

	//
	// Update binding & secret
	//

	switch rc.Connector.Spec.Deployment.DesiredState {
	case cos.DesiredStateReady:

		b, bs, bc, err := reify(&rc)
		if err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(rc.Connector, &bs, rc.M.GetScheme()); err != nil {
			return errors.Wrap(err, "unable to set binding secret controller reference")
		}
		if err := rc.PatchDependant(&bindingSecret, &bs); err != nil {
			return errors.Wrap(err, "unable to patch binding secret")
		}

		if err := controllerutil.SetControllerReference(rc.Connector, &bc, rc.M.GetScheme()); err != nil {
			return errors.Wrap(err, "unable to set binding config controller reference")
		}
		if err := rc.PatchDependant(&bindingConfig, &bc); err != nil {
			return errors.Wrap(err, "unable to patch binding config")
		}

		if err := controllerutil.SetControllerReference(rc.Connector, &b, rc.M.GetScheme()); err != nil {
			return errors.Wrap(err, "unable to set binding config controller reference")
		}
		if err := rc.PatchDependant(&binding, &b); err != nil {
			return errors.Wrap(err, "unable to patch binding")
		}

		conditions.SetReady(
			rc.Connector,
			metav1.ConditionTrue,
			conditions.ConditionReasonProvisioned,
			conditions.ConditionMessageProvisioned)

		rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
	case cos.DesiredStateStopped:
		conditions.SetReady(
			rc.Connector,
			metav1.ConditionFalse,
			conditions.ConditionReasonStopping,
			conditions.ConditionMessageStopping)

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
			conditions.SetReady(
				rc.Connector,
				metav1.ConditionFalse,
				conditions.ConditionReasonStopped,
				conditions.ConditionMessageStopped)

			rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
		}
	case cos.DesiredStateDeleted:
		conditions.SetReady(
			rc.Connector,
			metav1.ConditionFalse,
			conditions.ConditionReasonDeleting,
			conditions.ConditionMessageDeleting)

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
			conditions.SetReady(
				rc.Connector,
				metav1.ConditionFalse,
				conditions.ConditionReasonDeleted,
				conditions.ConditionMessageDeleted)

			rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
		}
	}

	return nil
}
