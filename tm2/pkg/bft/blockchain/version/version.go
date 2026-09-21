package bcver

import (
	types "github.com/gnolang/gno/tm2/pkg/bft/types/version"
)

// One of the six protocol-version constants that must move together; see
// RELEASING.md ("Protocol versions") and misc/release/bump-protocol-version.sh.
const Version = "v1.0.0-rc.0"

func init() {
	if types.BlockVersion != Version {
		panic("protocol version mismatch: types.BlockVersion is " + types.BlockVersion +
			" but bcver.Version is " + Version +
			"; the protocol-version constants must move together (see RELEASING.md)")
	}
}
