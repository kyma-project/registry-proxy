package state

import (
	"context"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnRemoveFinalizer(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if !controllerutil.RemoveFinalizer(&m.State.RegistryProxy, v1alpha1.Finalizer) {
		return requeue()
	}

	err := m.Client.Update(ctx, &m.State.RegistryProxy)
	return stopWithEventualError(err)
}
