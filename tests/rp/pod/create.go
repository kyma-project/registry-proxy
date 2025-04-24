package pod

import (
	"encoding/base64"
	"fmt"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/tests/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Create(utils *utils.TestUtils) error {
	rp, err := getRP(utils)
	if err != nil {
		return err
	}

	// docker-registry module credentials point to the docker-registry service, we ahve to convert that to our rp nodePort
	dockerCreds, err := getDockerCredentials(utils)
	if err != nil {
		return err
	}

	secret, err := fixSecret(utils, rp, dockerCreds)
	if err != nil {
		return err
	}
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

// TODO: common function
func getRP(utils *utils.TestUtils) (*v1alpha1.RegistryProxy, error) {
	var rp v1alpha1.RegistryProxy
	objectKey := client.ObjectKey{
		Name:      utils.RegistryProxyName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &rp); err != nil {
		return nil, err
	}

	return &rp, nil
}

func fixPod(utils *utils.TestUtils, rp *v1alpha1.RegistryProxy) (*v1.Pod, error) {
	podImage, err := getPodImage(utils, rp)
	if err != nil {
		return nil, err
	}
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.TestPod,
			Namespace: utils.Namespace,
		},
		Spec: v1.PodSpec{
			ImagePullSecrets: []v1.LocalObjectReference{
				{
					Name: utils.TestPod,
				},
			},
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: podImage,
				},
			},
		},
	}, nil
}

func getPodImage(utils *utils.TestUtils, rp *v1alpha1.RegistryProxy) (string, error) {
	nodeport := rp.Status.NodePort
	if nodeport == 0 {
		return "", fmt.Errorf("NodePort is not set in status")

	}
	return fmt.Sprintf("localhost:%d/%s", nodeport, utils.ImageName), nil
}

func getDockerCredentials(utils *utils.TestUtils) (*v1.Secret, error) {
	var secret v1.Secret
	objectKey := client.ObjectKey{
		Name:      "dockerregistry-config",
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

// we have to convert secret to kubernetes.io/dockerconfigjson
func fixSecret(utils *utils.TestUtils, rp *v1alpha1.RegistryProxy, dockerSecret *v1.Secret) (*v1.Secret, error) {
	username := string(dockerSecret.Data["username"])
	password := string(dockerSecret.Data["password"])
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	nodePort := rp.Status.NodePort

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

	return secret, nil
}
