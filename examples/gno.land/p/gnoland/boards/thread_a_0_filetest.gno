// PKGPATH: gno.land/r/demo/test
//
// Test creating a thread.
package test

import (
	"gno.land/p/gnoland/boards"
)

var board *boards.Board

func init() {
	board = boards.MustNew(1, perms{})
}

func main() {
	thread := boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")

	if thread.Board() == nil {
		panic("expected board to be assigned")
	}

	println(thread.Board().ID())
	println(thread.ID())
	println(thread.ID() == thread.ThreadID())
	println(thread.Creator())
	println(thread.OriginalBoardID())
	println(thread.Title)
	println(thread.Body)
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
// 1
// 1
// true
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// 0
// Title
// Body
