package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPeerAuthentication(t *testing.T) {
	t.Run("create peerAuthentication", func(t *testing.T) {
		c := minimalConnection()

		s := NewPeerAuthentication(c)

		require.NotNil(t, s)
		require.Equal(t, "test-c-name", s.GetName())
		require.Equal(t, "test-c-namespace", s.GetNamespace())
	})
}
