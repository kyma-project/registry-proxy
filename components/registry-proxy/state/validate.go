package state

import (
	"context"
	"fmt"
	"net/url"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnValidateReverseProxyURL(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	_, err := url.Parse(m.State.Connection.Spec.Proxy.URL)
	if err != nil {
		m.State.Connection.UpdateCondition(
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInvalidProxyURL,
			fmt.Sprintf("Invalid Connectivity Proxy URL: %s", err.Error()))
		return stop()
	}
	return nextState(sFnConnectivityProxyURL)
}
