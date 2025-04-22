package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryBoolCache(t *testing.T) {
	t.Run("set and get value", func(t *testing.T) {
		cache := NewInMemoryBoolCache()
		cache.Set(true)
		require.True(t, cache.Get())
	})
}
