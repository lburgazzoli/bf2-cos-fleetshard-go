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
	camel "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/conditions"
	meta2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/meta"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func ExtractConditions(conditions *[]metav1.Condition, binding camel.KameletBinding) error {

	gen, err := strconv.ParseInt(binding.Annotations[meta2.MetaDeploymentRevision], 10, 64)
	if err != nil {
		return errors.Wrap(err, "unable to determine revision")
	}

	// TODO: conditions must be filtered out
	for i := range binding.Status.Conditions {
		c := binding.Status.Conditions[i]

		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               "Workload" + string(c.Type),
			Status:             metav1.ConditionStatus(c.Status),
			LastTransitionTime: c.LastTransitionTime,
			Reason:             c.Reason,
			Message:            c.Message,

			// use ObservedGeneration to reference the deployment revision the
			// condition is about
			ObservedGeneration: gen,
		})
	}

	return nil
}

func ReadyCondition(connector cos.ManagedConnector) metav1.Condition {
	ready := metav1.Condition{
		Type:               conditions.ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		Reason:             conditions.ConditionReasonUnknown,
		Message:            conditions.ConditionMessageUnknown,
		ObservedGeneration: connector.Spec.Deployment.DeploymentResourceVersion,
	}

	if connector.Generation != connector.Status.ObservedGeneration {
		ready.Reason = conditions.ConditionMessageProvisioning
		ready.Message = conditions.ConditionReasonProvisioning
	}

	return ready
}

func SetReadyCondition(connector *cos.ManagedConnector, status metav1.ConditionStatus, reason string, message string) {
	controller.UpdateStatusCondition(&connector.Status.Conditions, conditions.ConditionTypeReady, func(condition *metav1.Condition) {
		condition.Status = status
		condition.Reason = reason
		condition.Message = message
	})
}
