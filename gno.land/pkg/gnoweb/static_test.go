//go:build !noembed

package gnoweb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The token stamped on asset URLs is derived from the content of the assets, not
// from when the process started: restarting with unchanged assets must not
// invalidate them for every client, and two replicas of the same binary must
// agree on the URL so a cache in front of them holds one entry per asset rather
// than one per replica.
func TestAssetsVersionIsDerivedFromAssetContent(t *testing.T) {
	t.Parallel()

	version := AssetsVersion()

	require.NotEmpty(t, version)
	assert.Equal(t, version, AssetsVersion(), "the version must not change between calls")
	assert.Regexp(t, `^[0-9a-f]+$`, version, "the version must be a content hash, not a timestamp")
	assert.True(t, strings.HasPrefix(strings.Trim(assetsHash, `"`), version),
		"the version and the asset ETag must come from the same content hash")
}
