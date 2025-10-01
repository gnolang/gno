// PKGPATH: gno.land/r/demo/test
//
// Test creating a reply to a reply.
package test

import (
	"std"

	"gno.land/p/gnoland/boards"
)

var parentReply *boards.Post

func init() {
	board := boards.MustNew(1, perms{})
	thread := boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
	err := thread.Save()
	if err != nil {
		panic(err)
	}

	parentReply = boards.MustNewReply(thread, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Message")
	err = parentReply.Save()
	if err != nil {
		panic(err)
	}
}

func main() {
	reply := boards.MustNewReply(parentReply, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Message")
	err := reply.Save()
	if err != nil {
		panic(err)
	}

	println("thread ID:", reply.ThreadID())
	println("parent reply ID:", parentReply.ID())
	println("reply ID:", reply.ID())
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
// thread ID: 1
// parent reply ID: 2
// reply ID: 3
