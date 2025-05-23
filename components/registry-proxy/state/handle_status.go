package state

import (
	"context"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnHandleStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// update ProxyURL & NodePort
	m.State.Connection.Status.ProxyURL = m.State.ProxyURL
	m.State.Connection.Status.NodePort = m.State.NodePort
	return nextState(nil)
}
