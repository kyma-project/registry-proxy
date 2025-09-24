package utils

import (
	"context"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetServedRegistryProxy(ctx context.Context, c client.Client) (*v1alpha1.RegistryProxy, error) {
	var registryProxyList v1alpha1.RegistryProxyList

	err := c.List(ctx, &registryProxyList)

	if err != nil {
		return nil, err
	}

	for _, item := range registryProxyList.Items {
		if !item.IsServedEmpty() && item.Status.Served == v1alpha1.ServedTrue {
			return &item, nil
		}
	}

	return nil, nil
}
