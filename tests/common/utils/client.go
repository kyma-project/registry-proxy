package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	operatorv1alpha1 "github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	rpv1alpha1 "github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetKuberentesClient() (client.Client, error) {
	config, err := LoadRestConfig("")
	if err != nil {
		return nil, err
	}

	err = operatorv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	err = rpv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{Scheme: scheme.Scheme})
}

func LoadRestConfig(context string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %w", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return loadRestConfigWithContext("", loadingRules, context)
}

func loadRestConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: api.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}
