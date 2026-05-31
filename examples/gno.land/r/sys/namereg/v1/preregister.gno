package namereg

import (
	susers "gno.land/r/sys/users"
)

// Pre-registered names bypass the Open Nym Tier `nym-...\d{3}` format
// intentionally. They are bootstrap names allocated at genesis and are
// not subject to the auto-registration regex. They DO appear in
// reservedNames as a defense-in-depth (so a future governance-bypass
// path can't accidentally reallocate them either), but at this layer
// they are written directly via susers.RegisterUserIgnoreCanonical.
//
// Uses the bypass path so curated confusables in the seed (e.g.
// `gnoland` and `gnolang` both canonicalize distinctly today, but a
// future addition that collides must not abort chain bring-up).
// Matches the genesis posture in r/sys/users/init.gno.

// pre-registered users
var preRegisteredUsers = []struct {
	Name    string
	Address address
}{
	// system names.
	// the goal is to make them either team/DAO-owned or ownerless.
	{"archive", "g1xlnyjrnf03ju82v0f98ruhpgnquk28knmjfe5k"}, // -> @archive
	{"demo", "g13ek2zz9qurzynzvssyc4sthwppnruhnp0gdz8n"},    // -> @demo
	{"gno", "g19602kd9tfxrfd60sgreadt9zvdyyuudcyxsz8a"},     // -> @gno
	{"gnoland", "g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"}, // -> @gnoland
	{"gnolang", "g1yjlnm3z2630gg5mryjd79907e0zx658wxs9hnd"}, // -> @gnolang
	{"gov", "g1g73v2anukg4ej7axwqpthsatzrxjsh0wk797da"},     // -> @gov
	{"nt", "g15ge0ae9077eh40erwrn2eq0xw6wupwqthpv34l"},      // -> @nt
	{"sys", "g1r929wt2qplfawe4lvqv9zuwfdcz4vxdun7qh8l"},     // -> @sys
	{"x", "g164sdpew3c2t3rvxj3kmfv7c7ujlvcw2punzzuz"},       // -> @x

	// test1 user
	{"test1", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}, // -> @test1
}

func init(cur realm) {
	// Add pre-registered users via the bypass path (genesis posture).
	// Names are curated but the underlying validation implementation could
	// potentially fail if changed and the list of curated names is not updated
	// accordingly, so registration errors are checked to make sure all
	// pre-registered names are effectibly registered, otherwise there's a chance
	// that an error could be silently ignored.
	for _, res := range preRegisteredUsers {
		err := susers.RegisterUserIgnoreCanonical(cross(cur), res.Name, res.Address)
		if err == nil {
			continue
		}

		// Only allow ErrNotWhitelisted on deploy for unit tests to work.
		// Error is triggered in tests because of block height != 0.
		if _, ok := err.(susers.ErrNotWhitelisted); !ok {
			panic(err)
		}
	}
}
