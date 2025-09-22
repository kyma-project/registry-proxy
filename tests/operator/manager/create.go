package manager

import (
	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/tests/common/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	rpObj := fixRegistryProxy(utils, utils.RegistryProxyName)

	return utils.Client.Create(utils.Ctx, rpObj)
}

func CreateSecond(utils *utils.TestUtils) error {
	rpObj := fixRegistryProxy(utils, utils.SecondRegistryProxyName)

	return utils.Client.Create(utils.Ctx, rpObj)
}

func fixRegistryProxy(testUtils *utils.TestUtils, name string) *v1alpha1.RegistryProxy {
	return &v1alpha1.RegistryProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testUtils.Namespace,
		},
		Spec: v1alpha1.RegistryProxySpec{},
	}
}
