package state

import (
	"context"
	"fmt"

	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	"github.tools.sap/kyma/registry-proxy/components/operator/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// sFnServedFilter checks if only one instance of RegistryProxy is running and disallows additional instances
func sFnServedFilter(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if m.State.RegistryProxy.IsServedEmpty() {
		err := setServedStatus(ctx, m)
		if err != nil {
			return stopWithEventualError(err)
		}
	}

	// instance is marked, we can now decide what to do with it
	if m.State.RegistryProxy.Status.Served == v1alpha1.ServedFalse {
		return stop()
	}
	return nextState(sFnAddFinalizer)
}

func setServedStatus(ctx context.Context, m *fsm.StateMachine) error {
	// check if there is any served instance
	servedRegistryProxy, err := utils.GetServedRegistryProxy(ctx, m.Client)
	if err != nil {
		return err
	}
	// no other Registry Proxy exists, mark this one os the served one
	if servedRegistryProxy == nil {
		m.State.RegistryProxy.Status.Served = v1alpha1.ServedTrue
		m.State.RegistryProxy.Status.State = v1alpha1.StateProcessing
	} else {
		// one served Registry Proxy already exists, add condition to this one and stop
		m.State.RegistryProxy.Status.Served = v1alpha1.ServedFalse
		m.State.RegistryProxy.Status.State = v1alpha1.StateWarning
		//nolint: staticcheck // linter is unhappy about the capital letter and a dot at the end
		err = fmt.Errorf("Only one instance of RegistryProxy is allowed (current served instance: %s/%s). This RegistryProxy CR is redundant. Remove it to fix the problem.",
			servedRegistryProxy.GetNamespace(), servedRegistryProxy.GetName())
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonRegistryProxyDuplicated,
			err.Error(),
		)
		return err
	}
	return nil
}
