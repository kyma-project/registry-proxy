package state

import (
	"context"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnAddFinalizer(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	instanceIsBeingDeleted := !m.State.RegistryProxy.GetDeletionTimestamp().IsZero()
	instanceHasFinalizer := controllerutil.ContainsFinalizer(&m.State.RegistryProxy, v1alpha1.Finalizer)
	if !instanceHasFinalizer {
		// in case instance has no finalizer and instance is being deleted - end reconciliation and allow for deletion by Kubernetes
		if instanceIsBeingDeleted {
			return stop()
		}

		// there is no finalizer and instance is not being deleted - add finalizer
		if err := addFinalizer(ctx, m); err != nil {
			return stopWithEventualError(err)
		}
	}
	return nextState(sFnInitialize)
}

func addFinalizer(ctx context.Context, m *fsm.StateMachine) error {
	// in case instance does not have finalizer - add it and update instance
	controllerutil.AddFinalizer(&m.State.RegistryProxy, v1alpha1.Finalizer)
	// update Registry Proxy right away
	return m.Client.Update(ctx, &m.State.RegistryProxy)
}
