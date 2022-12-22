package controller

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateStatusCondition(conditions *[]metav1.Condition, conditionType string, consumer func(*metav1.Condition)) {
	c := meta.FindStatusCondition(*conditions, conditionType)
	if c == nil {
		c = &metav1.Condition{
			Type: conditionType,
		}
	}

	consumer(c)

	meta.SetStatusCondition(conditions, *c)

}
