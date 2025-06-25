package connectivityproxy

import (
	"fmt"

	"github.tools.sap/kyma/registry-proxy/tests/common/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func VerifyMockStatefulSetDeletion(utils *utils.TestUtils) error {
	err := VerifyMockStatefulSet(utils)
	if err == nil {
		return fmt.Errorf("expected StatefulSet %s/connectivity-proxy to be deleted, but it still exists", utils.Namespace)
	}

	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func VerifyMockStatefulSet(utils *utils.TestUtils) error {
	statefulSet := fixStatefulSet(utils)
	err := utils.Client.Get(utils.Ctx, types.NamespacedName{
		Name:      statefulSet.Name,
		Namespace: statefulSet.Namespace,
	}, statefulSet)
	if err != nil {
		return err
	}

	if statefulSet.Status.AvailableReplicas < 1 {
		return fmt.Errorf("expected at least one available replica for StatefulSet %s/%s, got %d", statefulSet.Namespace, statefulSet.Name, statefulSet.Status.AvailableReplicas)
	}

	return nil
}
