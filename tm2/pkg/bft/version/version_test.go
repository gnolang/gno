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
//
// Note that the guards this covers are init() guards, so for the pairs they
// already cover the panic arrives during package initialisation and the
// assertions below never run — CI reports the drift as a panic naming the two
// constants rather than as a test failure. The assertions are what catch a
// constant that no init() guard happens to pair with.
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
// to peers actually carries a version. If an entry went empty on both sides,
// VersionSet.CompatibleWith would compare two empty majors, find them equal,
// and negotiate the component at the empty version — silently disabling the
// check rather than failing it. See TestVersionSetEmptyOnBothSidesNegotiatesNothing.
func TestVersionSetIsComplete(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"bft", "abci", "blockchain", "p2p"} {
		info, ok := VersionSet.Get(name)
		require.True(t, ok, "VersionSet is missing the %q entry", name)
		assert.NotEmpty(t, info.Version, "VersionSet entry %q has an empty version", name)
	}
}
