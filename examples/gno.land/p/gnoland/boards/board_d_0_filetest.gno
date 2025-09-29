// PKGPATH: gno.land/r/demo/test
//
// Test iterating threads.
package test

import (
	"std"

	"gno.land/p/gnoland/boards"
)

var board = boards.MustNew(1, perms{})

func init() {
	for i := 0; i < 3; i++ {
		thread := boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
		err := thread.Save()
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	println("threads:", board.ThreadsCount())
	board.IterateThreads(0, board.ThreadsCount(), func(thread *boards.Post) bool {
		println(thread.ID())
		return false
	})
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
// threads: 3
// 1
// 2
// 3
