package pod

import (
	"encoding/base64"
	"fmt"

	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/tests/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Create(utils *utils.TestUtils) error {
	iprp, err := getIPRP(utils)
	if err != nil {
		return err
	}

	// docker-registry module credentials point to the docker-registry service, we ahve to convert that to our iprp nodePort
	dockerCreds, err := getDockerCredentials(utils)
	if err != nil {
		return err
	}

	secret, err := fixSecret(utils, iprp, dockerCreds)
	if err != nil {
		return err
	}
	err = utils.Client.Create(utils.Ctx, secret)
	if err != nil {
		return err
	}

	pod, err := fixPod(utils, iprp)
	if err != nil {
		return err
	}

	return utils.Client.Create(utils.Ctx, pod)
}

// TODO: common function
func getIPRP(utils *utils.TestUtils) (*v1alpha1.ImagePullReverseProxy, error) {
	var iprp v1alpha1.ImagePullReverseProxy
	objectKey := client.ObjectKey{
		Name:      utils.ImagePullReverseProxyName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &iprp); err != nil {
		return nil, err
	}

	return &iprp, nil
}

func fixPod(utils *utils.TestUtils, iprp *v1alpha1.ImagePullReverseProxy) (*v1.Pod, error) {
	podImage, err := getPodImage(utils, iprp)
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

func getPodImage(utils *utils.TestUtils, iprp *v1alpha1.ImagePullReverseProxy) (string, error) {
	nodeport := iprp.Status.NodePort
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
func fixSecret(utils *utils.TestUtils, iprp *v1alpha1.ImagePullReverseProxy, dockerSecret *v1.Secret) (*v1.Secret, error) {
	username := string(dockerSecret.Data["username"])
	password := string(dockerSecret.Data["password"])
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	nodePort := iprp.Status.NodePort

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
