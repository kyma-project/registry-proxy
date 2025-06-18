package state

import (
	"context"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnValidateConnectivityProxyCRD(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if !m.Cache.Get() {
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConnectivityProxyCrdUnknown,
			"Connectivity Proxy not installed. This module is required.",
		)
		return requeueAfter(time.Minute)
	}
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionPrerequisitesSatisfied,
		metav1.ConditionTrue,
		v1alpha1.ConditionReasonConnectivityProxyCrdFound,
		"Connectivity Proxy installed.")
	return nextState(sFnApplyResources)
}
