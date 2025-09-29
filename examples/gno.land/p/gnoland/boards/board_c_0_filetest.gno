// PKGPATH: gno.land/r/demo/test
//
// Test deleting a thread from a board.
package test

import (
	"std"

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

func (perms) HasRole(std.Address, boards.Role) bool                                          { return false }
func (perms) HasPermission(std.Address, boards.Permission) bool                              { return false }
func (perms) SetUserRoles(realm, std.Address, ...boards.Role)                                {}
func (perms) RemoveUser(realm, std.Address) bool                                             { return false }
func (perms) HasUser(std.Address) bool                                                       { return false }
func (perms) UsersCount() int                                                                { return 0 }
func (perms) IterateUsers(start, count int, fn boards.UsersIterFn) bool                      { return false }
func (perms) WithPermission(realm, std.Address, boards.Permission, boards.Args, func(realm)) {}

// Output:
// false
