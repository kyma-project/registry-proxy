package state

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/chart"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// run registry-proxy chart installation
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
	m.State.FlagsBuilder.WithManagedByLabel("registry-proxy-operator")
	m.State.FlagsBuilder.WithIstioInstalled(m.IstioReadiness.Get())
	updateImages(m.State.FlagsBuilder)

	flags, err := m.State.FlagsBuilder.Build()
	if err != nil {
		m.Log.Warnf("error while building chart flags for resource %s: %s",
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

	// install component
	err = chart.Install(m.State.ChartConfig, flags)
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

func updateImages(fb chart.FlagsBuilder) {
	updateImageIfOverride("IMAGE_REGISTRY_PROXY", fb.WithImageRegistryProxy)
	updateImageIfOverride("IMAGE_CONNECTION", fb.WithImageConnection)
}

func updateImageIfOverride(envName string, updateFunction chart.ImageReplace) {
	imageName := os.Getenv(envName)
	if imageName != "" {
		updateFunction(imageName)
	}
}
