package state

import (
	"context"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"

	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnHandlePodStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// TODO: implement
	return stop()
	// /healthz - always true
	// /readyz - check / -> 200
}
