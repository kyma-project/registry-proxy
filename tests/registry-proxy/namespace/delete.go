package namespace

import "github.tools.sap/kyma/registry-proxy/tests/registry-proxy/utils"

func Delete(utils *utils.TestUtils) error {
	namespace := fixNamespace(utils)

	return utils.Client.Delete(utils.Ctx, namespace)
}
