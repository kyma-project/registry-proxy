package state

import (
	"context"
	"time"

	"github.com/kyma-project/manager-toolkit/installation/base/resource"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// delete Registry Proxy based on previously installed resources
func sFnDeleteResources(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.State.RegistryProxy.Status.State = v1alpha1.StateDeleting
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionTypeDeleted,
		metav1.ConditionUnknown,
		v1alpha1.ConditionReasonDeletion,
		"Uninstalling",
	)

	return nextState(sFnSafeDeletionState)
}

func sFnSafeDeletionState(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if err := chart.CheckCRDOrphanResources(m.State.ChartConfig); err != nil {
		// stop state machine with a warning and requeue reconciliation in 1min
		// warning state indicates that user intervention would fix it. Its not reconciliation error.
		m.State.RegistryProxy.Status.State = v1alpha1.StateWarning
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeletionErr,
			err.Error(),
		)
		return stopWithEventualError(err)
	}

	return deleteResources(m)
}

func deleteResources(m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	done, err := chart.Uninstall(m.State.ChartConfig, &chart.UninstallOpts{
		// first uninstall secrets to avoid issues with finalizers
		UninstallFirst: resource.HasKind("Secret"),
	})
	if err != nil {
		return uninstallResourcesError(m, err)
	}
	if !done {
		return awaitingResourcesRemoval(m)
	}

	m.State.RegistryProxy.Status.State = v1alpha1.StateDeleting
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionTypeDeleted,
		metav1.ConditionTrue,
		v1alpha1.ConditionReasonDeleted,
		"Registry Proxy module deleted",
	)

	// if resources are ready to be deleted, remove finalizer
	return nextState(sFnRemoveFinalizer)
}

func uninstallResourcesError(m *fsm.StateMachine, err error) (fsm.StateFn, *ctrl.Result, error) {
	m.Log.Warnf("error while uninstalling resource %s: %s",
		client.ObjectKeyFromObject(&m.State.RegistryProxy), err.Error())
	m.State.RegistryProxy.Status.State = v1alpha1.StateError
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionTypeDeleted,
		metav1.ConditionFalse,
		v1alpha1.ConditionReasonDeletionErr,
		err.Error(),
	)
	return stopWithEventualError(err)
}

func awaitingResourcesRemoval(m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.State.RegistryProxy.Status.State = v1alpha1.StateDeleting
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionTypeDeleted,
		metav1.ConditionTrue,
		v1alpha1.ConditionReasonDeletion,
		"Deleting module resources",
	)

	// wait one sec until ctrl-mngr remove finalizers from secrets
	return requeueAfter(time.Second)
}
