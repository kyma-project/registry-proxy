package state

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// verify if all workloads are in ready state
func sFnVerifyResources(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	result, err := chart.Verify(m.State.ChartConfig)
	if err != nil {
		m.Log.Warnf("error while verifying resource %s: %s",
			client.ObjectKeyFromObject(&m.State.RegistryProxy), err.Error())
		m.State.RegistryProxy.Status.State = v1alpha1.StateError

		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInstallationErr,
			err.Error(),
		)
		return stopWithEventualError(err)
	}

	if !result.Ready && result.Reason == chart.DeploymentVerificationProcessing {
		return requeueAfter(time.Second * 3)
	}

	if !result.Ready {
		// verification failed
		m.State.RegistryProxy.Status.State = v1alpha1.StateError
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionTypeDeploymentFailure,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonDeploymentReplicaFailure,
			result.Reason,
		)
		return stopWithEventualError(errors.New(result.Reason))
	}

	// remove possible previous DeploymentFailure condition
	m.State.RegistryProxy.RemoveCondition(v1alpha1.ConditionTypeDeploymentFailure)

	m.State.RegistryProxy.Status.State = v1alpha1.StateReady
	m.State.RegistryProxy.UpdateCondition(
		v1alpha1.ConditionTypeInstalled,
		metav1.ConditionTrue,
		v1alpha1.ConditionReasonInstalled,
		"Registry Proxy installed",
	)
	return stop()
}
