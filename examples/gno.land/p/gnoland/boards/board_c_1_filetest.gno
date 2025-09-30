// PKGPATH: gno.land/r/demo/test
//
// Test deleting a thread that doesn't exist from a board.
package test

import (
	"gno.land/p/gnoland/boards"
)

var board = boards.MustNew(1, perms{})

func main() {
	err := board.DeleteThread(404)
	if err == nil {
		panic("expected delete thread to return an error")
	}

	println(err.Error())
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
// thread not found, ID=404
