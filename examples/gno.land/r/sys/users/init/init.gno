// Package init provides basic user registration.
package init

import (
	"chain"

	"gno.land/r/sys/users"
)

// Bootstrap registers this package as a controller in r/sys/users.
func Bootstrap(_ realm) {
	users.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/users/init"))
}

// RegisterUser registers a new user in r/sys/users.
func RegisterUser(_ realm, name string, addr address) {
	if err := users.RegisterUser(cross, name, addr); err != nil {
		panic(err)
	}
}
