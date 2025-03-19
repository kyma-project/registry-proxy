package iprp

import (
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	iprpObj := fixImagePullReverseProxy(utils)

	return utils.Client.Create(utils.Ctx, iprpObj)
}

func fixImagePullReverseProxy(testUtils *utils.TestUtils) *v1alpha1.ImagePullReverseProxy {
	return &v1alpha1.ImagePullReverseProxy{
		ObjectMeta: v1.ObjectMeta{
			Name:      testUtils.ImagePullReverseProxyName,
			Namespace: testUtils.Namespace,
		},
		Spec: v1alpha1.ImagePullReverseProxySpec{
			ProxyURL:   testUtils.ProxyURL,
			TargetHost: testUtils.TargetHost,
			LogLevel:   "debug",
		},
	}
}
