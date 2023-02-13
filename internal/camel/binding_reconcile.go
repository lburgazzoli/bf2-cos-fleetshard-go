package camel

import (
	"fmt"
	kamelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/conditions"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/pointer"
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

	rc.Connector.Status.ObservedGeneration = rc.Connector.Generation
	rc.Connector.Status.ObservedDeploymentResourceVersion = rc.Connector.Spec.DeploymentResourceVersion

	switch rc.Connector.Spec.DesiredState {
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
		err := fmt.Errorf("unsupported desired state %s", rc.Connector.Spec.DesiredState)

		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(condition *cos.Condition) {
			condition.Status = metav1.ConditionFalse
			condition.Reason = conditions.ConditionReasonError
			condition.Message = err.Error()
		})

		return true, err
	}

	return true, nil
}

func handleReady(
	rc controller.ReconciliationContext,
	binding *kamelv1alpha1.KameletBinding,
	bindingSecret *corev1.Secret,
	bindingConfig *corev1.ConfigMap,
) error {

	c := conditions.Find(rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned)
	if c == nil {
		c = &cos.Condition{}
		c.Type = conditions.ConditionTypeProvisioned
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonProvisioning
		c.Message = conditions.ConditionReasonProvisioning
	}

	c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
	c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion

	b, bs, bc, err := reify(&rc)
	if err != nil {
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonError
		c.Message = err.Error()

		conditions.Set(&rc.Connector.Status.Conditions, *c)

		return err
	}

	if err := patchDependant(rc, bindingSecret, &bs); err != nil {
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonError
		c.Message = err.Error()

		conditions.Set(&rc.Connector.Status.Conditions, *c)

		return errors.Wrap(err, "unable to reconcile binding secrete")
	}

	if err := patchDependant(rc, bindingConfig, &bc); err != nil {
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonError
		c.Message = err.Error()

		conditions.Set(&rc.Connector.Status.Conditions, *c)

		return errors.Wrap(err, "unable to reconcile binding config")
	}

	if err := patchDependant(rc, binding, &b); err != nil {
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonError
		c.Message = err.Error()

		conditions.Set(&rc.Connector.Status.Conditions, *c)

		return errors.Wrap(err, "unable to reconcile binding")
	}

	c.Status = metav1.ConditionTrue
	c.Reason = conditions.ConditionReasonProvisioned
	c.Message = conditions.ConditionMessageProvisioned

	conditions.Set(&rc.Connector.Status.Conditions, *c)

	return nil
}

// TODO: should scale the binding
func handleStop(
	rc controller.ReconciliationContext,
	binding *kamelv1alpha1.KameletBinding,
	bindingSecret *corev1.Secret,
	bindingConfig *corev1.ConfigMap,
) error {

	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(c *cos.Condition) {
		c.Status = metav1.ConditionTrue
		c.Reason = conditions.ConditionReasonProvisioned
		c.Message = conditions.ConditionMessageProvisioned
		c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
		c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
	})
	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(c *cos.Condition) {
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonStopping
		c.Message = conditions.ConditionReasonStopping
		c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
		c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
	})

	b := binding.DeepCopy()
	b.Spec.Replicas = pointer.Of(int32(0))

	if err := patchDependant(rc, binding, b); err != nil {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(c *cos.Condition) {
			c.Status = metav1.ConditionFalse
			c.Reason = conditions.ConditionReasonError
			c.Message = err.Error()
		})

		return errors.Wrap(err, "unable to scale binding")
	}

	if b.Status.Replicas != nil && *b.Status.Replicas == 0 {

		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(c *cos.Condition) {
			c.Status = metav1.ConditionFalse
			c.Reason = conditions.ConditionReasonStopped
			c.Message = conditions.ConditionReasonStopped
			c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
			c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
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

	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeProvisioned, func(c *cos.Condition) {
		c.Status = metav1.ConditionTrue
		c.Reason = conditions.ConditionReasonProvisioned
		c.Message = conditions.ConditionMessageProvisioned
		c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
		c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
	})
	conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(c *cos.Condition) {
		c.Status = metav1.ConditionFalse
		c.Reason = conditions.ConditionReasonDeleting
		c.Message = conditions.ConditionReasonDeleting
		c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
		c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
	})

	deleted := 0

	for _, r := range []client.Object{binding, bindingSecret, bindingConfig} {
		if err := rc.DeleteDependant(r); err != nil {
			if k8serrors.IsNotFound(err) {
				deleted++
			} else {
				conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(c *cos.Condition) {
					c.Status = metav1.ConditionFalse
					c.Reason = conditions.ConditionReasonError
					c.Message = err.Error()
					c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
					c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
				})
				return err
			}
		}
	}

	if deleted == 3 {
		conditions.Update(&rc.Connector.Status.Conditions, conditions.ConditionTypeReady, func(c *cos.Condition) {
			c.Status = metav1.ConditionFalse
			c.Reason = conditions.ConditionReasonDeleted
			c.Message = conditions.ConditionReasonDeleted
			c.ObservedGeneration = rc.Connector.Status.ObservedGeneration
			c.ResourceRevision = rc.Connector.Status.ObservedDeploymentResourceVersion
		})
	}

	return nil
}
