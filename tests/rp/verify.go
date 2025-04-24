package rp

import (
	"fmt"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/tests/rp/deployment"
	"github.tools.sap/kyma/registry-proxy/tests/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDeletion(utils *utils.TestUtils) error {
	err := Verify(utils)
	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func Verify(utils *utils.TestUtils) error {
	var rp v1alpha1.RegistryProxy
	objectKey := client.ObjectKey{
		Name:      utils.RegistryProxyName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &rp); err != nil {
		return err
	}

	if err := verifyState(utils, &rp); err != nil {
		return err
	}

	if err := verifyStatus(&rp); err != nil {
		return err
	}

	return deployment.VerifyEnvs(utils, &rp)
}

// check if all data from the spec is reflected in the status
func verifyStatus(rp *v1alpha1.RegistryProxy) error {
	status := rp.Status
	spec := rp.Spec

	if err := isSpecValueReflectedInStatus(spec.ProxyURL, status.ProxyURL); err != nil {
		return err
	}

	if status.NodePort == 0 {
		return fmt.Errorf("NodePort is not set in status")
	}

	return nil
}

func isSpecValueReflectedInStatus(specValue string, statusValue string) error {
	if specValue == "" {
		// value is not set in the spec, so value in the status may be empty or defauled
		return nil
	}

	if specValue != statusValue {
		return fmt.Errorf("value '%s' not found in status", specValue)
	}

	return nil
}

func verifyState(utils *utils.TestUtils, rp *v1alpha1.RegistryProxy) error {
	for _, condition := range rp.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionReady) {
			if condition.Reason == string(v1alpha1.ConditionReasonProbeSuccess) &&
				condition.Status == metav1.ConditionTrue &&
				condition.Message == "" {
				return nil
			}
			return fmt.Errorf("ConditionReady is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionReady not found")
}
