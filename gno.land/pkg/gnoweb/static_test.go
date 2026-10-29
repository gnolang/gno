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

// The feature controllers that open a code editor share one copy of CodeMirror
// (a chunk-*.js built with esbuild's --splitting) and the BaseController of
// the controller.js every page loads, instead of each carrying its own.
func TestFeatureControllersShareChunks(t *testing.T) {
	t.Parallel()

	entries, err := assets.ReadDir("public/js")
	require.NoError(t, err)

	var editors []string
	for _, e := range entries {
		b, err := assets.ReadFile("public/js/" + e.Name())
		require.NoError(t, err)
		if strings.Contains(string(b), "cm-editor") {
			editors = append(editors, e.Name())
		}
	}
	require.Len(t, editors, 1, "CodeMirror must be in one file, got %v", editors)
	assert.Regexp(t, `^chunk-[A-Z0-9]+\.js$`, editors[0])

	for _, name := range []string{"controller-playground.js", "controller-run.js"} {
		b, err := assets.ReadFile("public/js/" + name)
		require.NoError(t, err)
		assert.Contains(t, string(b), `from"./`+editors[0]+`"`, name)
		assert.Contains(t, string(b), `from"./controller.js"`, name)
		assert.NotContains(t, string(b), "[data-action]", name+" must not carry its own BaseController")
	}
}
