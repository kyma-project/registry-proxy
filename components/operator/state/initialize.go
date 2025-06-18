package state

import (
	"context"

	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// in case instance is being deleted and has finalizer - delete all resources
	instanceIsBeingDeleted := !m.State.RegistryProxy.GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nextState(sFnDeleteResources)
	}

	// TODO: install resources
	return nextState(sFnValidateConnectivityProxyCRD)
}
