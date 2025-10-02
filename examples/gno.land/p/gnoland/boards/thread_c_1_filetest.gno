// PKGPATH: gno.land/r/demo/test
//
// Test deleting a reply that doesn't exist from a thread.
package test

import (
	"gno.land/p/gnoland/boards"
)

var thread *boards.Post

func init() {
	board := boards.MustNew(1, perms{})
	thread = boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
}

func main() {
	err := thread.DeleteReply(404)
	if err == nil {
		panic("expected an error")
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
// reply not found
