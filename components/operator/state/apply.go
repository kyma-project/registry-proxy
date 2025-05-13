package state

import (
	"context"
	"fmt"

	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/chart"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// run serverless chart installation
func sFnApplyResources(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// set condition Installed if it does not exist
	if !m.State.RegistryProxy.IsConditionSet(v1alpha1.ConditionTypeInstalled) {
		m.State.RegistryProxy.Status.State = v1alpha1.StateProcessing
		m.State.RegistryProxy.UpdateCondition(
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	}

	// update common labels for all rendered resources
	flags := map[string]interface{}{
		"global": map[string]interface{}{
			"commonLabels": map[string]interface{}{
				"app.kubernetes.io/managed-by": "registry-proxy-operator",
			},
		},
	}

	// install component
	err := chart.Install(m.State.ChartConfig, flags)
	if err != nil {
		fmt.Println(err)
		m.Log.Warnf("error while installing resource %s: %s",
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

	// switch state verify
	return nextState(sFnVerifyResources)
}
