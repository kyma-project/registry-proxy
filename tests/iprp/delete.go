package iprp

import "github.tools.sap/kyma/image-pull-reverse-proxy/tests/utils"

func Delete(utils *utils.TestUtils) error {
	iprp := fixImagePullReverseProxy(utils)

	return utils.Client.Delete(utils.Ctx, iprp)
}
