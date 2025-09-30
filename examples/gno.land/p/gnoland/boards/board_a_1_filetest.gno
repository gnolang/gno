// PKGPATH: gno.land/r/demo/test
//
// Test board members iteration.
package test

import (
	"gno.land/p/gnoland/boards"
)

var members = []boards.User{
	{
		Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		Roles:   []boards.Role{"a", "b"},
	},
	{
		Address: "g1e5hud66rs7ye4zgeqmqvwfhurs2mnf408hdqcl",
		Roles:   []boards.Role{"c"},
	},
}

func main() {
	board := boards.MustNew(1, perms{})

	println(board.MembersCount())

	stopped := board.IterateMembers(0, board.MembersCount(), func(addr address, roles []boards.Role) bool {
		println(addr)
		for _, r := range roles {
			println("role", string(r))
		}
		return false
	})
	println(stopped)
}

type perms struct{}

func (perms) HasRole(address, boards.Role) bool                                          { return false }
func (perms) HasPermission(address, boards.Permission) bool                              { return false }
func (perms) SetUserRoles(realm, address, ...boards.Role)                                {}
func (perms) RemoveUser(realm, address) bool                                             { return false }
func (perms) HasUser(address) bool                                                       { return false }
func (perms) UsersCount() int                                                            { return len(members) }
func (perms) WithPermission(realm, address, boards.Permission, boards.Args, func(realm)) {}

func (perms) IterateUsers(start, count int, fn boards.UsersIterFn) bool {
	for _, u := range members {
		fn(u)
	}

	return true
}

// Output:
// 2
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// role a
// role b
// g1e5hud66rs7ye4zgeqmqvwfhurs2mnf408hdqcl
// role c
// true
