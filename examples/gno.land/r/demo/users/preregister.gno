package users

import (
	"std"

	"gno.land/p/demo/users"
)

// pre-restricted names
var preRestrictedNames = []string{
	"bitcoin", "cosmos", "newtendermint", "administrator", "ethereum",
}

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

	// famous name
	{"test1", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},   // -> @test1
	{"jaekwon", "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"}, // -> @jae
	{"manfred", "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"}, // -> @moul
}

func init() {
	// add pre-registered users
	for _, res := range preRegisteredUsers {
		// assert not already registered.
		_, ok := name2User.Get(res.Name)
		if ok {
			panic("name already registered")
		}

		_, ok = addr2User.Get(res.Address.String())
		if ok {
			panic("address already registered")
		}

		counter++
		user := &users.User{
			Address: res.Address,
			Name:    res.Name,
			Profile: "",
			Number:  counter,
			Invites: int(0),
			Inviter: admin,
		}
		name2User.Set(res.Name, user)
		addr2User.Set(res.Address.String(), user)
	}

	// add pre-restricted names
	for _, name := range preRestrictedNames {
		if _, ok := name2User.Get(name); ok {
			panic("name already registered")
		}

		restricted.Set(name, true)
	}
}
