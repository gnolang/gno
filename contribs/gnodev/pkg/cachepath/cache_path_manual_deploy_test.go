package cachepath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachePath(t *testing.T) {

	assert.NotNil(t, cache, "Cache should be initialized")
	assert.NotNil(t, cache.data, "Cache data map should be initialized")

	t.Run("Set and Get", func(t *testing.T) {
		path := "gno.land/r/test/example"

		exists := Get(path)
		assert.False(t, exists, "Path should not exist before setting")

		Set(path)

		exists = Get(path)
		assert.True(t, exists, "Path should exist after setting")
	})

	t.Run("Multiple paths", func(t *testing.T) {

		cache.data = make(map[string]bool)

		paths := []string{
			"gno.land/r/test/path1",
			"gno.land/r/test/path2",
			"gno.land/p/demo/path3",
		}

		for _, path := range paths {
			Set(path)
		}

		for _, path := range paths {
			exists := Get(path)
			assert.True(t, exists, "Path %s should exist after setting", path)
		}

		nonExistentPath := "gno.land/r/test/nonexistent"
		exists := Get(nonExistentPath)
		assert.False(t, exists, "Non-existent path should not exist")
	})
}
