package users

import (
	"std"

	susers "gno.land/p/sys/users"
	sys "gno.land/r/sys/users"
)

// pre-registered users
var preRegisteredUsers = []struct {
	Name    string
	Address std.Address
}{
	// system name
	{"archives", "g1xlnyjrnf03ju82v0f98ruhpgnquk28knmjfe5k"}, // -> @r_archives
	{"demo", "g13ek2zz9qurzynzvssyc4sthwppnruhnp0gdz8n"},     // -> @r_demo
	{"gno", "g19602kd9tfxrfd60sgreadt9zvdyyuudcyxsz8a"},      // -> @r_gno
	{"gnoland", "g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"},  // -> @r_gnoland
	{"gnolang", "g1yjlnm3z2630gg5mryjd79907e0zx658wxs9hnd"},  // -> @r_gnolang
	{"gov", "g1g73v2anukg4ej7axwqpthsatzrxjsh0wk797da"},      // -> @r_gov
	{"nt", "g15ge0ae9077eh40erwrn2eq0xw6wupwqthpv34l"},       // -> @r_nt
	{"sys", "g1r929wt2qplfawe4lvqv9zuwfdcz4vxdun7qh8l"},      // -> @r_sys
	{"x", "g164sdpew3c2t3rvxj3kmfv7c7ujlvcw2punzzuz"},        // -> @r_x

	// test1 user
	{"test1", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}, // -> @test1
}

func init() {
	height := std.GetHeight()

	// add pre-registered users
	for _, res := range preRegisteredUsers {
		base := &susers.User{
			Address: res.Address,
			Name:    res.Name,
		}

		if err := sys.WriteUser(base); err != nil {
			panic(err)
		}

		extendedUserStore.Set(res.Address.String(), &User{
			base:         base,
			Bio:          "Preregistered user.",
			Ordinal:      sys.UserCount(),
			RegisteredAt: height,
		})
	}
}
