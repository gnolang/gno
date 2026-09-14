package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abciver "github.com/gnolang/gno/tm2/pkg/bft/abci/version"
	bcver "github.com/gnolang/gno/tm2/pkg/bft/blockchain/version"
	blockver "github.com/gnolang/gno/tm2/pkg/bft/types/version"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	p2pver "github.com/gnolang/gno/tm2/pkg/p2p/version"
)

// TestProtocolVersionsAgree pins the invariant that the init() guards in this
// package, bcver and blockver enforce at startup: the six protocol-version
// constants are one version, spread across six files for import-cycle reasons
// only. A bump that misses one of them panics every node at init, which is a
// bad way to find out. misc/release/bump-protocol-version.sh moves all six
// together; this test is what tells CI it did.
func TestProtocolVersionsAgree(t *testing.T) {
	t.Parallel()

	versions := map[string]string{
		"tm2/pkg/bft/version.Version":            Version,
		"tm2/pkg/bft/abci/version.Version":       abciver.Version,
		"tm2/pkg/bft/blockchain/version.Version": bcver.Version,
		"tm2/pkg/bft/types/version.BlockVersion": blockver.BlockVersion,
		"tm2/pkg/p2p/version.Version":            p2pver.Version,
		"tm2/pkg/crypto.Version":                 crypto.Version,
	}

	for name, got := range versions {
		assert.Equal(t, Version, got,
			"%s disagrees with tm2/pkg/bft/version.Version; run misc/release/bump-protocol-version.sh", name)
	}
}

// TestVersionSetIsComplete checks that every component this package advertises
// to peers actually carries a version. An entry with an empty version makes
// VersionSet.CompatibleWith compare empty majors, which matches anything —
// silently disabling the negotiation rather than failing it.
func TestVersionSetIsComplete(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"bft", "abci", "blockchain", "p2p"} {
		info, ok := VersionSet.Get(name)
		require.True(t, ok, "VersionSet is missing the %q entry", name)
		assert.NotEmpty(t, info.Version, "VersionSet entry %q has an empty version", name)
	}
}
