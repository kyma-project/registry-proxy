package pod

import (
	"errors"
	"github.tools.sap/kyma/registry-proxy/tests/common/utils"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Verify(utils *utils.TestUtils) error {
	pod, err := getPod(utils)
	if err != nil {
		return err
	}

	return verifyPod(pod)

}

func getPod(utils *utils.TestUtils) (*v1.Pod, error) {
	var pod v1.Pod
	objectKey := client.ObjectKey{
		Name:      utils.TestPod,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

func verifyPod(pod *v1.Pod) error {
	if pod.Status.Phase != v1.PodRunning {
		return errors.New("pod is not ready")
	}

	return nil
}
