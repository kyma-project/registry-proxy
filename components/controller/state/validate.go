package state

import (
	"context"
	"fmt"
	"net/url"

	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

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
