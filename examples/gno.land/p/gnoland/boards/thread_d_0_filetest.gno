// PKGPATH: gno.land/r/demo/test
//
// Test reposting a thread into another board.
package test

import (
	"gno.land/p/gnoland/boards"
)

var (
	board  *boards.Board
	thread *boards.Post
)

func init() {
	origBoard := boards.MustNew(1, perms{})
	thread = boards.MustNewThread(origBoard, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
	err := thread.Save()
	if err != nil {
		panic(err)
	}

	board = boards.MustNew(2, perms{})
}

func main() {
	repost := boards.MustNewRepost(thread, board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	if repost.Board() == nil {
		panic("expected destination board to be assigned")
	}

	println("Original Board ID:", repost.OriginalBoardID())
	println("Original Thread ID:", repost.ParentID())
	println("Repost Board ID:", repost.Board().ID())
	println("Repost ID:", repost.ID())
	println(repost.ID() == thread.ThreadID())
	println(repost.Creator())
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
// Original Board ID: 1
// Original Thread ID: 1
// Repost Board ID: 2
// Repost ID: 1
// true
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
