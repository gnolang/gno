package version

import (
	abciver "github.com/gnolang/gno/tm2/pkg/bft/abci/version"
	bcver "github.com/gnolang/gno/tm2/pkg/bft/blockchain/version"
	p2pver "github.com/gnolang/gno/tm2/pkg/p2p/version"
	verset "github.com/gnolang/gno/tm2/pkg/versionset"
)

var (
	// The major or minor versions must bump when components bump.
	// The TendermintClassic software version
	Version    = "v1.0.0-rc.0"
	VersionSet verset.VersionSet
)

func init() {
	// Defensive checks
	//nolint:goconst
	if abciver.Version != "v1.0.0-rc.0" ||
		bcver.Version != "v1.0.0-rc.0" ||
		p2pver.Version != "v1.0.0-rc.0" {
		panic("bump Version")
	}

	VersionSet.Set(verset.VersionInfo{
		Name:    "bft",
		Version: Version,
	})
	VersionSet.Set(verset.VersionInfo{
		Name:    "abci",
		Version: abciver.Version,
	})
	VersionSet.Set(verset.VersionInfo{
		Name:    "blockchain",
		Version: bcver.Version,
	})
	VersionSet.Set(verset.VersionInfo{
		Name:    "p2p",
		Version: p2pver.Version,
	})
}
