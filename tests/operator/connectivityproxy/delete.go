package connectivityproxy

import "github.tools.sap/kyma/registry-proxy/tests/common/utils"

func DeleteMockStatefulSet(utils *utils.TestUtils) error {
	statefulSet := fixStatefulSet(utils)
	return utils.Client.Delete(utils.Ctx, statefulSet)
}
