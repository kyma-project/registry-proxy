package connection

import (
	"github.com/kyma-project/registry-proxy/tests/common/utils"
)

func Delete(utils *utils.TestUtils) error {
	rp := fixConnection(utils)

	return utils.Client.Delete(utils.Ctx, rp)
}
