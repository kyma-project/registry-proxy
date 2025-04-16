package rp

import "github.tools.sap/kyma/registry-proxy/tests/utils"

func Delete(utils *utils.TestUtils) error {
	rp := fixImagePullReverseProxy(utils)

	return utils.Client.Delete(utils.Ctx, rp)
}
