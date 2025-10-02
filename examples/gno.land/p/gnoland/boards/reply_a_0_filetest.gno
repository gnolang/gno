// PKGPATH: gno.land/r/demo/test
//
// Test creating a thread reply.
package test

import (
	"gno.land/p/gnoland/boards"
)

var thread *boards.Post

func init() {
	board := boards.MustNew(1, perms{})
	thread = boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
	err := thread.Save()
	if err != nil {
		panic(err)
	}
}

func main() {
	reply := boards.MustNewReply(thread, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Message")
	err := reply.Save()
	if err != nil {
		panic(err)
	}

	println("thread ID:", reply.ThreadID())
	println("reply ID:", reply.ID())
	println(reply.Creator())
	println(reply.Body)
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
// thread ID: 1
// reply ID: 2
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Message
