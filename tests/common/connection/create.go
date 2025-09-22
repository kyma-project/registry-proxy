package connection

import (
	"encoding/base64"
	"fmt"

	"github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/tests/common/connection/dockerregistry"
	"github.com/kyma-project/registry-proxy/tests/common/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	authTokenSecretName = "auth-token-secret"
)

func Create(utils *utils.TestUtils) error {
	connectionObj := fixConnection(utils)

	if utils.AuthToken {
		dockerCreds, err := dockerregistry.GetDockerCredentials(utils.Ctx, utils.Client, utils.Namespace)
		if err != nil {
			return err
		}
		secret := fixSecret(utils, dockerCreds)
		err = utils.Client.Create(utils.Ctx, secret)
		if err != nil {
			return err
		}

		connectionObj.Spec.Target.Authorization.HeaderSecret = authTokenSecretName
	}

	return utils.Client.Create(utils.Ctx, connectionObj)
}

func fixConnection(testUtils *utils.TestUtils) *v1alpha1.Connection {
	connection := &v1alpha1.Connection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testUtils.ConnectionName,
			Namespace: testUtils.Namespace,
		},
		Spec: v1alpha1.ConnectionSpec{
			Proxy: v1alpha1.ConnectionSpecProxy{
				URL: testUtils.ProxyURL,
			},
			Target: v1alpha1.ConnectionSpecTarget{
				Host: testUtils.TargetHost,
			},

			LogLevel: "debug",
		},
	}

	return connection
}

func fixSecret(utils *utils.TestUtils, dockerSecret *corev1.Secret) *corev1.Secret {
	username := string(dockerSecret.Data["username"])
	password := string(dockerSecret.Data["password"])
	authEncoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	auth := fmt.Sprintf("Basic %s", authEncoded)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authTokenSecretName,
			Namespace: utils.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"authorizationHeader": []byte(auth),
		},
	}

	return secret
}
