package state

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnValidateConnectivityProxyCRD(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if m.State.RegistryProxy.Spec.Proxy.URL != "" {
		_, err := url.Parse(m.State.RegistryProxy.Spec.Proxy.URL)
		if err != nil {
			m.State.RegistryProxy.UpdateCondition(
				v1alpha1.ConditionPrerequisitesSatisfied,
				metav1.ConditionFalse,
				v1alpha1.ConditionReasonProxyURLInavlid,
				fmt.Sprintf("Invalid Proxy URL: %s", err.Error()))
			return stop()
		}
		// skip Connectivity Proxy check, user has set proxy URL manually
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConnectivityProxySkipped,
			"Connectivity Proxy check skipped, .spec.proxy.url is set.")
		return nextState(sFnApplyResources)
	}
	if !m.ConnectivityProxyReadiness.Get() {
		m.State.RegistryProxy.Status.State = v1alpha1.StateWarning
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConnectivityProxyUnavailable,
			"Connectivity Proxy is unavailable. This module is required.",
		)
		return requeueAfter(time.Minute)
	}
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionPrerequisitesSatisfied,
		metav1.ConditionTrue,
		v1alpha1.ConditionReasonConnectivityProxyAvailable,
		"Connectivity Proxy installed.")
	return nextState(sFnApplyResources)
}
