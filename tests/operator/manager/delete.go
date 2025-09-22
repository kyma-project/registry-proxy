package manager

import (
	"github.com/kyma-project/registry-proxy/tests/common/utils"
)

func Delete(utils *utils.TestUtils) error {
	rp := fixRegistryProxy(utils, utils.RegistryProxyName)

	return utils.Client.Delete(utils.Ctx, rp)
}

func DeleteSecond(utils *utils.TestUtils) error {
	rp := fixRegistryProxy(utils, utils.SecondRegistryProxyName)

	return utils.Client.Delete(utils.Ctx, rp)
}
