package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Run("create service", func(t *testing.T) {
		c := minimalConnection()

		s := NewService(c)

		require.NotNil(t, s)
		require.Equal(t, "test-c-name", s.GetName())
		require.Equal(t, "test-c-namespace", s.GetNamespace())
	})

	t.Run("create service with desired NodePort", func(t *testing.T) {
		rp := minimalRegistryProxyWithPort(3001)

		s := NewService(rp)

		require.NotNil(t, s)
		require.Equal(t, "test-rp-name", s.GetName())
		require.Equal(t, "test-rp-namespace", s.GetNamespace())
		require.Equal(t, int32(3001), s.Spec.Ports[0].NodePort)
	})
}
