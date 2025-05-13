package chart

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_ ManifestCache = (*inMemoryManifestCache)(nil)
	_ ManifestCache = (*secretManifestCache)(nil)
)

var (
	emptyRegistryProxySpecManifest = RegistryProxySpecManifest{}
)

type ManifestCache interface {
	Set(context.Context, client.ObjectKey, RegistryProxySpecManifest) error
	Get(context.Context, client.ObjectKey) (RegistryProxySpecManifest, error)
	Delete(context.Context, client.ObjectKey) error
}

// inMemoryManifestCache provides an in-memory processor to store registry proxy Spec and rendered chart manifest. By using sync.Map for caching,
// concurrent operations to the processor from diverse reconciliations are considered safe.
//
// Inside the processor is stored chart manifest with used custom flags by client.ObjectKey key.
type inMemoryManifestCache struct {
	processor sync.Map
}

// NewInMemoryManifestCache returns a new instance of inMemoryManifestCache.
func NewInMemoryManifestCache() *inMemoryManifestCache {
	return &inMemoryManifestCache{
		processor: sync.Map{},
	}
}

// Get loads the RegistryProxySpecManifest from inMemoryManifestCache for the passed client.ObjectKey.
func (r *inMemoryManifestCache) Get(_ context.Context, key client.ObjectKey) (RegistryProxySpecManifest, error) {
	value, ok := r.processor.Load(key)
	if !ok {
		return emptyRegistryProxySpecManifest, nil
	}

	return *value.(*RegistryProxySpecManifest), nil
}

// Set saves the passed flags and manifest into inMemoryManifestCache for the client.ObjectKey.
func (r *inMemoryManifestCache) Set(_ context.Context, key client.ObjectKey, spec RegistryProxySpecManifest) error {
	r.processor.Store(key, &spec)

	return nil
}

// Delete deletes flags and manifest from inMemoryManifestCache for the passed client.ObjectKey.
func (r *inMemoryManifestCache) Delete(_ context.Context, key client.ObjectKey) error {
	r.processor.Delete(key)
	return nil
}

// secretManifestCache - provides an Secret based processor to store registry proxy Spec and rendered chart manifest.
//
// Inside the secret we store manifest and flags used to render it.
type secretManifestCache struct {
	client client.Client
}

type RegistryProxySpecManifest struct {
	ManagerUID  string
	CustomFlags map[string]interface{}
	Manifest    string
}

// NewSecretManifestCache - returns a new instance of SecretManifestCache.
func NewSecretManifestCache(client client.Client) *secretManifestCache {
	return &secretManifestCache{
		client: client,
	}
}

// Delete - removes Secret cache based on the passed client.ObjectKey.
func (m *secretManifestCache) Delete(ctx context.Context, key client.ObjectKey) error {
	err := m.client.Delete(ctx, &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
	})

	return client.IgnoreNotFound(err)
}

// Get - loads the RegistryProxySpecManifest from SecretManifestCache based on the passed client.ObjectKey.
func (m *secretManifestCache) Get(ctx context.Context, key client.ObjectKey) (RegistryProxySpecManifest, error) {
	secret := corev1.Secret{}
	err := m.client.Get(ctx, key, &secret)
	if errors.IsNotFound(err) {
		return emptyRegistryProxySpecManifest, nil
	}
	if err != nil {
		return emptyRegistryProxySpecManifest, err
	}

	spec := RegistryProxySpecManifest{}
	err = json.Unmarshal(secret.Data["spec"], &spec)
	if err != nil {
		return emptyRegistryProxySpecManifest, err
	}

	return spec, nil
}

// Set - saves the passed flags and manifest into Secret based on the client.ObjectKey.
func (m *secretManifestCache) Set(ctx context.Context, key client.ObjectKey, spec RegistryProxySpecManifest) error {
	byteSpec, err := json.Marshal(&spec)
	if err != nil {
		return err
	}

	secret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Data: map[string][]byte{
			"spec": byteSpec,
		},
	}

	err = m.client.Update(ctx, &secret)
	if !errors.IsNotFound(err) {
		return err
	}

	return m.client.Create(ctx, &secret)
}
