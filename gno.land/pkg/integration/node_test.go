package integration

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/jaekwon/testify/require"
)

func TestNewInMemory(t *testing.T) {
	logger := log.TestingLogger()

	node, err := NewNode(logger, NodeConfig{})
	require.NoError(t, err)
	require.NotNil(t, node)
}
