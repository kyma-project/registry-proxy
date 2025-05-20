package manager

import (
	"fmt"

	"github.tools.sap/kyma/registry-proxy/tests/common/utils"

	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
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

func VerifyDeletionStuck(utils *utils.TestUtils) error {
	registryProxy, err := getRegistryProxy(utils, utils.RegistryProxyName)
	if err != nil {
		return err
	}

	return verifyDeletionStuck(&registryProxy)
}

func Verify(utils *utils.TestUtils) error {
	registryProxy, err := getRegistryProxy(utils, utils.RegistryProxyName)
	if err != nil {
		return err
	}

	return verifyState(&registryProxy)
}

func VerifyStuck(utils *utils.TestUtils) error {
	registryProxy, err := getRegistryProxy(utils, utils.SecondRegistryProxyName)
	if err != nil {
		return err
	}

	return verifyStateStuck(&registryProxy)
}

func getRegistryProxy(utils *utils.TestUtils, name string) (v1alpha1.RegistryProxy, error) {
	var registryProxy v1alpha1.RegistryProxy
	objectKey := client.ObjectKey{
		Name:      name,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &registryProxy); err != nil {
		return v1alpha1.RegistryProxy{}, err
	}
	return registryProxy, nil
}

func verifyState(rp *v1alpha1.RegistryProxy) error {
	for _, condition := range rp.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionTypeInstalled) {
			if condition.Reason == string(v1alpha1.ConditionReasonInstalled) &&
				condition.Status == metav1.ConditionTrue &&
				condition.Message == "Registry Proxy installed" {
				return nil
			}
			return fmt.Errorf("ConditionReady is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionReady not found")
}

func verifyStateStuck(rp *v1alpha1.RegistryProxy) error {
	for _, condition := range rp.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionTypeConfigured) {
			if condition.Reason == string(v1alpha1.ConditionReasonRegistryProxyDuplicated) &&
				condition.Status == metav1.ConditionFalse &&
				condition.Message == fmt.Sprintf("Only one instance of RegistryProxy is allowed (current served instance: %s/test-registry-proxy). This RegistryProxy CR is redundant. Remove it to fix the problem.", rp.Namespace) {
				return nil
			}
			return fmt.Errorf("ConditionConfigured is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionConfigured not found")
}

func verifyDeletionStuck(rp *v1alpha1.RegistryProxy) error {
	for _, condition := range rp.Status.Conditions {
		if condition.Type == string(v1alpha1.ConditionTypeDeleted) {
			if condition.Reason == string(v1alpha1.ConditionReasonDeletionErr) &&
				condition.Status == metav1.ConditionFalse &&
				condition.Message == "found 1 items with VersionKind registry-proxy.kyma-project.io/v1alpha1" {
				return nil
			}
			return fmt.Errorf("ConditionDeleted is not in expected state: %v", condition)
		}
	}
	return fmt.Errorf("ConditionDeleted not found")
}
