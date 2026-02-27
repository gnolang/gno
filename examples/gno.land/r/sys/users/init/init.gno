// Package init provides user registration for dev and test chains.
// It is pre-whitelisted as a controller in r/sys/users.
package init

import "gno.land/r/sys/users"

// RegisterUser registers a new user in r/sys/users.
func RegisterUser(_ realm, name string, addr address) {
	if err := users.RegisterUser(cross, name, addr); err != nil {
		panic(err)
	}
}
