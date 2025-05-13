package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *RegistryProxy) IsConditionSet(conditionType ConditionType) bool {
	return meta.FindStatusCondition(
		s.Status.Conditions, string(conditionType),
	) != nil
}

func (s *RegistryProxy) IsConditionTrue(conditionType ConditionType) bool {
	condition := meta.FindStatusCondition(s.Status.Conditions, string(conditionType))
	return condition != nil && condition.Status == metav1.ConditionTrue
}
