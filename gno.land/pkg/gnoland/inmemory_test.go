package gnoland

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/jaekwon/testify/require"
)

func TestNewInMemory(t *testing.T) {
	logger := log.TestingLogger()

	node, err := NewInMemory(logger, InMemoryConfig{})
	require.NoError(t, err)
	require.NotNil(t, node)
}
