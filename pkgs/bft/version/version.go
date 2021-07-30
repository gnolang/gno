package version

import (
	abciver "github.com/gnolang/gno/pkgs/bft/abci/version"
	bcver "github.com/gnolang/gno/pkgs/bft/blockchain/version"
	p2pver "github.com/gnolang/gno/pkgs/p2p/version"
)

var (
	// The major or minor versions must bump when components bump.
	// The TendermintClassic software version
	Version = "v1.0.0-rc.0"
)

func init() {
	if abciver.Version != "v1.0.0-rc.0" ||
		bcver.Version != "v1.0.0-rc.0" ||
		p2pver.Version != "v1.0.0-rc.0" {
		panic("bump Version")
	}
}
