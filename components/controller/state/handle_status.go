package state

import (
	"context"

	"github.tools.sap/kyma/registry-proxy/components/controller/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnHandleStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// update ProxyURL & NodePort
	m.State.ReverseProxy.Status.ProxyURL = m.State.ProxyURL
	m.State.ReverseProxy.Status.NodePort = m.State.NodePort
	return nextState(nil)
}
