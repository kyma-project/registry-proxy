package state

import (
	"context"
	"fmt"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func sFnValidateConnectivityProxyCRD(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if !m.Cache.Get() {
		m.State.ReverseProxy.UpdateCondition(
			v1alpha1.ConditionConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConnectivityProxyCrdUnknownn,
			fmt.Sprintf("Connectivity Proxy not installed. This module is required. "),
		)
		return requeueAfter(time.Minute)
	}
	m.State.ReverseProxy.UpdateCondition(
		v1alpha1.ConditionConfigured,
		metav1.ConditionTrue,
		v1alpha1.ConditionReasonConnectivityProxyCrdFound,
		"Connectivity Proxy installed.")
	return nextState(sFnValidateReverseProxyURL)
}

func sFnValidateReverseProxyURL(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	_, err := url.Parse(m.State.ReverseProxy.Spec.ProxyURL)
	if err != nil {
		m.State.ReverseProxy.UpdateCondition(
			v1alpha1.ConditionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInvalidProxyURL,
			fmt.Sprintf("Invalid Connectivity Proxy URL: %s", err.Error()))
		return stop()
	}
	return nextState(sFnConnectivityProxyURL)
}
