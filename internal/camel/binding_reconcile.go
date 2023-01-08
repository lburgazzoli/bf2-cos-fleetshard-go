package camel

import (
	"fmt"
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

func Apply(rc controller.ReconciliationContext) (bool, error) {

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
		return false, errors.Wrap(err, "failure loading dependant KameletBinding")
	}
	if err := rc.GetDependant(&bindingSecret); err != nil {
		return false, errors.Wrap(err, "failure loading dependant KameletBinding secret")
	}
	if err := rc.GetDependant(&bindingConfig); err != nil {
		return false, errors.Wrap(err, "failure loading dependant KameletBinding config")
	}

	//
	// Update connector conditions
	//

	if err := extractDependantConditions(&rc.Connector.Status.Conditions, binding); err != nil {
		return false, errors.Wrap(err, "unable to compute binding conditions")
	}

	//
	// Update binding & secret
	//

	switch rc.Connector.Spec.Deployment.DesiredState {
	case cos.DesiredStateReady:
		if err := handleReady(rc, &binding, &bindingSecret, &bindingConfig); err != nil {
			return true, err
		}
	case cos.DesiredStateStopped:
		if err := handleStop(rc, &binding, &bindingSecret, &bindingConfig); err != nil {
			return true, err
		}
	case cos.DesiredStateDeleted:
		if err := handleDelete(rc, &binding, &bindingSecret, &bindingConfig); err != nil {
			return true, err
		}
	default:
		err := fmt.Errorf("unsupported desired state %s", rc.Connector.Spec.Deployment.DesiredState)

		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonError
			condition.Message = err.Error()
		})

		return true, err
	}

	rc.Connector.Status.ObservedGeneration = rc.Connector.Generation

	for i := range rc.Connector.Status.Conditions {
		rc.Connector.Status.Conditions[i].ObservedGeneration = rc.Connector.Status.ObservedGeneration
		rc.Connector.Status.Conditions[i].ResourceRevision = rc.Connector.Spec.Deployment.DeploymentResourceVersion
	}

	return true, nil
}

func handleReady(
	rc controller.ReconciliationContext,
	binding *kamelv1alpha1.KameletBinding,
	bindingSecret *corev1.Secret,
	bindingConfig *corev1.ConfigMap,
) error {

	// TODO: add methods to reduce conditions handling duplication

	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
		condition.Status = metav1.ConditionFalse
		condition.Reason = conditions.ConditionReasonProvisioning
		condition.Message = conditions.ConditionReasonProvisioning
	})

	b, bs, bc, err := reify(&rc)
	if err != nil {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonError
			condition.Message = err.Error()
		})

		return err
	}

	if err := patchDependant(rc, bindingSecret, &bs); err != nil {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonError
			condition.Message = err.Error()
		})

		return errors.Wrap(err, "unable to reconcile binding secrete")
	}

	if err := patchDependant(rc, bindingConfig, &bc); err != nil {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonError
			condition.Message = err.Error()
		})

		return errors.Wrap(err, "unable to reconcile binding config")
	}

	if err := patchDependant(rc, binding, &b); err != nil {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonError
			condition.Message = err.Error()
		})

		return errors.Wrap(err, "unable to reconcile binding")
	}

	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
		condition.Status = metav1.ConditionTrue
		condition.Reason = conditions.ConditionReasonProvisioned
		condition.Message = conditions.ConditionMessageProvisioned
	})

	return nil
}

func handleStop(
	rc controller.ReconciliationContext,
	binding *kamelv1alpha1.KameletBinding,
	bindingSecret *corev1.Secret,
	bindingConfig *corev1.ConfigMap,
) error {

	// TODO: add methods to reduce conditions handling duplication

	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
		condition.Status = metav1.ConditionFalse
		condition.Reason = conditions.ConditionReasonStopping
		condition.Message = conditions.ConditionMessageStopping
	})

	deleted := 0

	for _, r := range []client.Object{binding, bindingSecret, bindingConfig} {
		if err := rc.DeleteDependant(r); err != nil {
			if k8serrors.IsNotFound(err) {
				deleted++
			} else {
				conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
					condition.Status = metav1.ConditionFalse
					condition.Reason = conditions.ConditionReasonError
					condition.Message = err.Error()
				})

				return err
			}
		}
	}

	if deleted == 3 {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = conditions.ConditionReasonStopped
			condition.Message = conditions.ConditionMessageStopped
		})
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonStopped
			condition.Message = conditions.ConditionMessageStopped
		})
	}

	return nil
}

func handleDelete(
	rc controller.ReconciliationContext,
	binding *kamelv1alpha1.KameletBinding,
	bindingSecret *corev1.Secret,
	bindingConfig *corev1.ConfigMap,
) error {

	// TODO: add methods to reduce conditions handling duplication

	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
		condition.Status = metav1.ConditionTrue
		condition.Reason = conditions.ConditionReasonDeleting
		condition.Message = conditions.ConditionMessageDeleting
	})

	deleted := 0

	for _, r := range []client.Object{binding, bindingSecret, bindingConfig} {
		if err := rc.DeleteDependant(r); err != nil {
			if k8serrors.IsNotFound(err) {
				deleted++
			} else {
				conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
					condition.Status = metav1.ConditionFalse
					condition.Reason = conditions.ConditionReasonError
					condition.Message = err.Error()
				})
				return err
			}
		}
	}

	if deleted == 3 {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionTrue
			condition.Reason = conditions.ConditionReasonDeleted
			condition.Message = conditions.ConditionMessageDeleted
		})
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonDeleted
			condition.Message = conditions.ConditionReasonDeleted
		})
	}

	return nil
}
