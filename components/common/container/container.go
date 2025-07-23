package container

import (
	corev1 "k8s.io/api/core/v1"
)

// Get retrieves a container by name from a slice of containers.
func Get(containers []corev1.Container, name string) *corev1.Container {
	for _, container := range containers {
		if container.Name == name {
			return &container
		}
	}
	return nil
}
