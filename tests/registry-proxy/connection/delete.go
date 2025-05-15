package connection

import "github.tools.sap/kyma/registry-proxy/tests/registry-proxy/utils"

func Delete(utils *utils.TestUtils) error {
	rp := fixConnection(utils)

	return utils.Client.Delete(utils.Ctx, rp)
}
