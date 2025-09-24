package namespace

import (
	"github.com/kyma-project/registry-proxy/tests/common/utils"
)

func Delete(utils *utils.TestUtils) error {
	namespace := fixNamespace(utils)

	return utils.Client.Delete(utils.Ctx, namespace)
}
