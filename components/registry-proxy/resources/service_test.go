package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Run("create service", func(t *testing.T) {
		rp := minimalRegistryProxy()

		s := NewService(rp)

		require.NotNil(t, s)
		require.Equal(t, "test-rp-name", s.GetName())
		require.Equal(t, "test-rp-namespace", s.GetNamespace())
	})
}
