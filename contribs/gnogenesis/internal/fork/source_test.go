package fork

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openSourceTest wraps openSource with the default workersPerEndpoint so
// tests focus on URL parsing rather than concurrency tuning.
func openSourceTest(s string) (Source, error) {
	return openSource(s, defaultWorkersPerEndpoint)
}

func TestOpenSource_SingleURL(t *testing.T) {
	t.Parallel()

	src, err := openSourceTest("http://localhost:26657")
	require.NoError(t, err)
	defer src.Close()

	rs, ok := src.(*rpcSource)
	require.True(t, ok, "expected *rpcSource, got %T", src)
	assert.Equal(t, []string{"http://localhost:26657"}, rs.rpcURLs)
	assert.Len(t, rs.clients, 1)
}

func TestOpenSource_MultipleURLs(t *testing.T) {
	t.Parallel()

	src, err := openSourceTest("http://a:26657,http://b:26657,http://c:26657")
	require.NoError(t, err)
	defer src.Close()

	rs, ok := src.(*rpcSource)
	require.True(t, ok)
	assert.Equal(t, []string{"http://a:26657", "http://b:26657", "http://c:26657"}, rs.rpcURLs)
	assert.Len(t, rs.clients, 3)
}

func TestOpenSource_TrimsWhitespaceAndSkipsEmpty(t *testing.T) {
	t.Parallel()

	src, err := openSourceTest("  http://a:26657 , , http://b:26657  ")
	require.NoError(t, err)
	defer src.Close()

	rs, ok := src.(*rpcSource)
	require.True(t, ok)
	assert.Equal(t, []string{"http://a:26657", "http://b:26657"}, rs.rpcURLs)
}

func TestOpenSource_RejectsMixedURLAndPath(t *testing.T) {
	t.Parallel()

	_, err := openSourceTest("http://a:26657,/some/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http")
}

func TestOpenSource_RejectsAllEmpty(t *testing.T) {
	t.Parallel()

	_, err := openSourceTest(",,")
	require.Error(t, err)
}

func TestOpenSource_RejectsNonHTTPSchemeInList(t *testing.T) {
	t.Parallel()

	for _, s := range []string{
		"ws://localhost:26657,http://b:26657",
		"http://a:26657,tcp://b:26657",
		"http://a:26657,ftp://example.org",
	} {
		_, err := openSourceTest(s)
		require.Error(t, err, "input: %s", s)
		assert.Contains(t, err.Error(), "http")
	}
}
