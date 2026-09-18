package version

import (
	abciver "github.com/gnolang/gno/tm2/pkg/bft/abci/version"
	bcver "github.com/gnolang/gno/tm2/pkg/bft/blockchain/version"
	p2pver "github.com/gnolang/gno/tm2/pkg/p2p/version"
	verset "github.com/gnolang/gno/tm2/pkg/versionset"
)

var (
	// The major or minor versions must bump when components bump.
	// The TendermintClassic software version.
	//
	// One of the six protocol-version constants that must move together; see
	// RELEASING.md ("Protocol versions") and misc/release/bump-protocol-version.sh.
	//
	// This is not the release version. A release is identified by its git tag,
	// compiled in as tm2/pkg/version.Version, and that is what a halt proposal's
	// halt_min_version compares against. This constant is the protocol version
	// negotiated with peers: VersionSet.CompatibleWith rejects a peer whose major
	// differs, so bumping it partitions the network and is its own flag day.
	Version    = "v1.0.0-rc.0"
	VersionSet verset.VersionSet
)

func init() {
	// Defensive checks. Compare against the constant rather than a repeated
	// literal, so a bump is an edit to the constants alone.
	for _, component := range [...]struct{ name, version string }{
		{"abci", abciver.Version},
		{"blockchain", bcver.Version},
		{"p2p", p2pver.Version},
	} {
		if component.version != Version {
			panic("protocol version mismatch: bft is " + Version +
				" but " + component.name + " is " + component.version +
				"; the protocol-version constants must move together (see RELEASING.md)")
		}
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
