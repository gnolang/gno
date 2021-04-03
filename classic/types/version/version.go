package version

import (
	"github.com/tendermint/classic/crypto"
)

// Versioning for just the structure of the block.
const BlockVersion string = "v1.0.0-rc.0"

func init() {
	if crypto.Version != "v1.0.0-rc.0" {
		panic("bump BlockVersion?")
	}
}
