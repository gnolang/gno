// PKGPATH: gno.land/r/demo/test
//
// Test reposting a thread into another board using an invalid creator address.
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
	boards.MustNewRepost(thread, board, "foo")
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
// invalid thread repost creator address: foo
