package users

import (
	susers "gno.land/r/sys/users"
)

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

func init() {
	// add pre-registered users
	for _, res := range preRegisteredUsers {
		// Try registering, skip if it fails
		susers.RegisterUser(cross, res.Name, res.Address)
	}
}
