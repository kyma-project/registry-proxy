package pod

import (
	"encoding/base64"
	"fmt"

	"github.com/kyma-project/registry-proxy/tests/common/connection/dockerregistry"
	"github.com/kyma-project/registry-proxy/tests/common/utils"

	"github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Create(utils *utils.TestUtils) error {
	rp, err := getConnection(utils)
	if err != nil {
		return err
	}

	// docker-registry module credentials point to the docker-registry service, we have to convert that to our connection nodePort
	dockerCreds, err := dockerregistry.GetDockerCredentials(utils.Ctx, utils.Client, utils.Namespace)
	if err != nil {
		return err
	}

	secret := fixSecret(utils, rp, dockerCreds)

	err = utils.Client.Create(utils.Ctx, secret)
	if err != nil {
		return err
	}

	pod, err := fixPod(utils, rp)
	if err != nil {
		return err
	}

	return utils.Client.Create(utils.Ctx, pod)
}

func getConnection(utils *utils.TestUtils) (*v1alpha1.Connection, error) {
	var connection v1alpha1.Connection
	objectKey := client.ObjectKey{
		Name:      utils.ConnectionName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &connection); err != nil {
		return nil, err
	}

	return &connection, nil
}

func fixPod(utils *utils.TestUtils, connection *v1alpha1.Connection) (*v1.Pod, error) {
	podImage, err := getPodImage(utils, connection)
	if err != nil {
		return nil, err
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.TestPod,
			Namespace: utils.Namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: podImage,
				},
			},
		},
	}

	// use standard secret if we're not explicitly pushing hardcoded auth header
	if !utils.AuthToken {
		pod.Spec.ImagePullSecrets = []v1.LocalObjectReference{
			{
				Name: utils.TestPod,
			},
		}
	}

	return pod, nil
}

func getPodImage(utils *utils.TestUtils, connection *v1alpha1.Connection) (string, error) {
	nodeport := connection.Status.NodePort
	if nodeport == 0 {
		return "", fmt.Errorf("NodePort is not set in status")

	}
	return fmt.Sprintf("localhost:%d/%s", nodeport, utils.TaggedImageName), nil
}

// we have to convert secret to kubernetes.io/dockerconfigjson
func fixSecret(utils *utils.TestUtils, connection *v1alpha1.Connection, dockerSecret *v1.Secret) *v1.Secret {
	username := string(dockerSecret.Data["username"])
	password := string(dockerSecret.Data["password"])
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	nodePort := connection.Status.NodePort

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.TestPod,
			Namespace: utils.Namespace,
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": []byte(fmt.Sprintf(
				`{"auths":{"localhost:%d":{"username":"%s","password":"%s","email":"example@sap.com","auth":"%s"}}}`, nodePort, username, password, auth),
			),
		},
	}

	return secret
}
