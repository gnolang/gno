// PKGPATH: gno.land/r/demo/test
//
// Test getting a thread that doesnt exists from a board.
package test

import (
	"gno.land/p/gnoland/boards"
)

var board = boards.MustNew(1, perms{})

func main() {
	_, found := board.GetThread(404)
	if found {
		panic("expected no thread to be found")
	}

	println("ok")
}

type perms struct{}

func (perms) HasRole(address, boards.Role) bool                                          { return false }
func (perms) HasPermission(address, boards.Permission) bool                              { return false }
func (perms) SetUserRoles(realm, address, ...boards.Role)                                {}
func (perms) RemoveUser(realm, address) bool                                             { return false }
func (perms) HasUser(address) bool                                                       { return false }
func (perms) UsersCount() int                                                            { return 0 }
func (perms) IterateUsers(start, count int, fn boards.UsersIterFn) bool                  { return false }
func (perms) WithPermission(realm, address, boards.Permission, boards.Args, func(realm)) {}

// Output:
// ok
