package dockerregistry

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDockerCredentials(ctx context.Context, k8sclient client.Client, namespace string) (*v1.Secret, error) {
	var secret v1.Secret
	objectKey := client.ObjectKey{
		Name:      "dockerregistry-config",
		Namespace: "kyma-system",
	}

	if err := k8sclient.Get(ctx, objectKey, &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}
