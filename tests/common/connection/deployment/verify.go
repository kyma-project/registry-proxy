package deployment

import (
	"fmt"

	"github.com/kyma-project/registry-proxy/tests/common/utils"

	connection "github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyEnvs(utils *utils.TestUtils, connection *connection.Connection) error {
	var deploy appsv1.Deployment
	objectKey := client.ObjectKey{
		Name:      utils.ConnectionName,
		Namespace: utils.Namespace,
	}

	err := utils.Client.Get(utils.Ctx, objectKey, &deploy)
	if err != nil {
		return err
	}

	return verifyDeployEnvs(&deploy, connection)
}

func verifyDeployEnvs(deploy *appsv1.Deployment, connection *connection.Connection) error {
	expectedEnvs := []corev1.EnvVar{
		{
			Name:  "PROXY_URL",
			Value: connection.Status.ProxyURL,
		},
		{
			Name:  "TARGET_HOST",
			Value: connection.Spec.Target.Host,
		},
	}

	for _, expectedEnv := range expectedEnvs {
		if !isEnvReflected(expectedEnv, &deploy.Spec.Template.Spec.Containers[0]) {
			return fmt.Errorf("env '%s' with value '%s' not found in deployment", expectedEnv.Name, expectedEnv.Value)
		}
	}

	return nil
}

func isEnvReflected(expected corev1.EnvVar, in *corev1.Container) bool {
	if expected.Value == "" {
		// return true if value is not overrided
		return true
	}

	for _, env := range in.Env {
		if env.Name == expected.Name {
			// return true if value is the same
			return env.Value == expected.Value
		}
	}

	return false
}
