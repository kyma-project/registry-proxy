package namespace

import "github.tools.sap/kyma/image-pull-reverse-proxy/tests/utils"

func Delete(utils *utils.TestUtils) error {
	namespace := fixNamespace(utils)

	return utils.Client.Delete(utils.Ctx, namespace)
}
