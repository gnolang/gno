// PKGPATH: gno.land/r/demo/test
//
// Test deleting a reply of a reply from a thread.
package test

import (
	"gno.land/p/gnoland/boards"
)

var (
	thread      *boards.Post
	parentReply *boards.Post
	replyID     boards.ID
)

func init() {
	board := boards.MustNew(1, perms{})
	thread = boards.MustNewThread(board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Title", "Body")
	err := thread.Save()
	if err != nil {
		panic(err)
	}

	parentReply = boards.MustNewReply(thread, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Message")
	err = parentReply.Save()
	if err != nil {
		panic(err)
	}

	reply := boards.MustNewReply(parentReply, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "Message")
	err = reply.Save()
	if err != nil {
		panic(err)
	}

	replyID = reply.ID()
}

func main() {
	err := thread.DeleteReply(replyID)
	if err != nil {
		panic(err)
	}

	_, found := thread.GetReply(replyID)
	println(found)

	_, found = parentReply.GetReply(replyID)
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
// false
