package state

import (
	"time"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"

	ctrl "sigs.k8s.io/controller-runtime"
)

// nolint:unused
var requeueResult = &ctrl.Result{
	Requeue: true,
}

func nextState(next fsm.StateFn) (fsm.StateFn, *ctrl.Result, error) {
	return next, nil, nil
}

// nolint:unused
func requeue() (fsm.StateFn, *ctrl.Result, error) {
	return nil, requeueResult, nil
}

// nolint:unparam
func requeueAfter(duration time.Duration) (fsm.StateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}

func stop() (fsm.StateFn, *ctrl.Result, error) {
	return nil, nil, nil
}

func stopWithEventualError(err error) (fsm.StateFn, *ctrl.Result, error) {
	return nil, nil, err
}

// nolint:unused
func stopWithErrorOrRequeue(err error) (fsm.StateFn, *ctrl.Result, error) {
	return nil, requeueResult, err
}

func StartState() fsm.StateFn {
	return sFnValidateReverseProxyURL
}
