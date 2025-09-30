// PKGPATH: gno.land/r/demo/test
//
// Test deleting a thread from a board.
package test

import (
	"gno.land/p/gnoland/boards"
)

var (
	thread *boards.Post
	board  = boards.MustNew(1, perms{})
)

func init() {
	thread = boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
	err := thread.Save()
	if err != nil {
		panic(err)
	}
}

func main() {
	err := board.DeleteThread(thread.ID())
	if err != nil {
		panic(err)
	}

	_, found := board.GetThread(thread.ID())
	println(found)
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
// false
