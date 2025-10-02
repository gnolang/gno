// PKGPATH: gno.land/r/demo/test
//
// Test creating a thread with an invalid owner address.
package test

import (
	"gno.land/p/gnoland/boards"
)

var board *boards.Board

func init() {
	board = boards.MustNew(1, perms{})
}

func main() {
	boards.MustNewThread(board, "foo", "Title", "Body")
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

// Error:
// invalid thread creator address: foo
