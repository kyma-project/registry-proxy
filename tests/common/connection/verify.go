package connection

import (
	"fmt"

	"github.com/kyma-project/registry-proxy/tests/common/utils"

	"github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/tests/common/connection/deployment"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDeletion(utils *utils.TestUtils) error {
	err := Verify(utils)
	if err == nil {
		return fmt.Errorf("expected error during deletion, got none")
	}
	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func Verify(utils *utils.TestUtils) error {
	var connection v1alpha1.Connection
	objectKey := client.ObjectKey{
		Name:      utils.ConnectionName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &connection); err != nil {
		return err
	}

	if err := verifyState(&connection); err != nil {
		return err
	}

	if err := verifyStatus(&connection); err != nil {
		return err
	}

	return deployment.VerifyEnvs(utils, &connection)
}

// check if all data from the spec is reflected in the status
func verifyStatus(connection *v1alpha1.Connection) error {
	status := connection.Status
	spec := connection.Spec

	if err := isSpecValueReflectedInStatus(spec.Proxy.URL, status.ProxyURL); err != nil {
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

func verifyState(rp *v1alpha1.Connection) error {
	for _, condition := range rp.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionConnectionReady) {
			if condition.Reason == string(v1alpha1.ConditionReasonEstablished) &&
				condition.Status == metav1.ConditionTrue &&
				condition.Message == "Target registry reachable" {
				return nil
			}
			return fmt.Errorf("ConditionConnectionReady is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionConnectionReady not found")
}
