package version

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Versioning for just the structure of the block.
//
// One of the six protocol-version constants that must move together; see
// RELEASING.md ("Protocol versions") and misc/release/bump-protocol-version.sh.
const BlockVersion string = "v1.0.0-rc.0"

func init() {
	// Compare against the constant rather than a repeated literal: bumping a
	// protocol version should be an edit to the constants alone, not a hunt for
	// every literal that mirrors them.
	if crypto.Version != BlockVersion {
		panic("protocol version mismatch: crypto.Version is " + crypto.Version +
			" but BlockVersion is " + BlockVersion +
			"; the protocol-version constants must move together (see RELEASING.md)")
	}
}
