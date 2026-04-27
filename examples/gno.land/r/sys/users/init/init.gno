// Package init provides basic user registration.
//
// SECURITY: every public function in this package is genesis-only. The
// realm exists to seed initial users into r/sys/users at chain genesis
// and to register itself as a controller. After genesis (block height
// > 0), no caller may use these functions.
//
// Without the genesis-only gate, the wide-open `RegisterUser` wrapper
// would let any EOA land-grab any name (including reserved-sounding
// names like "administrator" or "vitalik") to any address — for free,
// without the namereg/v1 payment, blacklist, or canonical-collision
// checks. The gate closes that bypass.
//
// Post-genesis user registration must go through a whitelisted
// controller that enforces its own policy (e.g. r/sys/namereg/v1).
package init

import (
	"chain"
	"chain/runtime"

	"gno.land/r/sys/users"
)

// Bootstrap registers this package as a controller in r/sys/users.
// Genesis-only via AddControllerAtGenesis's own height==0 gate.
func Bootstrap(_ realm) {
	users.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/users/init"))
}

// RegisterUser registers a new user in r/sys/users at chain genesis.
// PANICS if called after genesis (height > 0).
//
// Uses RegisterUserIgnoreCanonical: the genesis seed set is curated, and
// any deliberate confusable reservations (e.g. registering both `vitalik`
// and `vital1k` to two different addresses) must not abort chain bring-
// up. Decision #14 (later-wins) applies, so order in genesis_txs.jsonl
// determines which name owns the canonical pointer when stems collide.
func RegisterUser(_ realm, name string, addr address) {
	if runtime.ChainHeight() != 0 {
		panic("r/sys/users/init.RegisterUser: genesis-only; use a whitelisted controller post-genesis (e.g. r/sys/namereg/v1.Register)")
	}
	if err := users.RegisterUserIgnoreCanonical(cross, name, addr); err != nil {
		panic(err)
	}
}
