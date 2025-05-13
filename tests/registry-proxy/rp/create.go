package rp

import (
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/tests/registry-proxy/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	rpObj := fixRegistryProxy(utils)

	return utils.Client.Create(utils.Ctx, rpObj)
}

func fixRegistryProxy(testUtils *utils.TestUtils) *v1alpha1.Connection {
	return &v1alpha1.Connection{
		ObjectMeta: v1.ObjectMeta{
			Name:      testUtils.RegistryProxyName,
			Namespace: testUtils.Namespace,
		},
		Spec: v1alpha1.ConnectionSpec{
			ProxyURL:   testUtils.ProxyURL,
			TargetHost: testUtils.TargetHost,
			LogLevel:   "debug",
		},
	}
}
