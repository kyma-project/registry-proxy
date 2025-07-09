package reverseproxy

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	t.Run("should return a new reverse proxy", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		proxy, err := New(":1234", "http://connectivity.proxy", "target", "id", log)
		require.NoError(t, err)
		require.NotNil(t, proxy)
	})
	t.Run("return error on invalid connectivityProxyURL", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		_, err := New(":1234", ":invalid", "target", "id", log)
		require.Error(t, err)
	})
}
