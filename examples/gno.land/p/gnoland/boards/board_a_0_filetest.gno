// PKGPATH: gno.land/r/demo/test
//
// Test board creation and default values.
package test

import (
	"gno.land/p/gnoland/boards"
)

var options = []boards.Option{
	boards.WithName("foo"),
	boards.WithCreator("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
}

func main() {
	board := boards.MustNew(1, perms{}, options...)

	println(board.ID())
	println(board.Name())
	println(board.Aliases())
	println(board.Readonly())
	println(board.Creator())
	println(board.CreatedAt())
	println(board.MembersCount())
	println(board.ThreadsCount())
}

type perms struct{}

func (perms) HasRole(address, boards.Role) bool                                          { return false }
func (perms) HasPermission(address, boards.Permission) bool                              { return false }
func (perms) SetUserRoles(realm, address, ...boards.Role)                                {}
func (perms) RemoveUser(realm, address) bool                                             { return false }
func (perms) HasUser(address) bool                                                       { return false }
func (perms) UsersCount() int                                                            { return 42 }
func (perms) IterateUsers(start, count int, fn boards.UsersIterFn) bool                  { return false }
func (perms) WithPermission(realm, address, boards.Permission, boards.Args, func(realm)) {}

// Output:
// 1
// foo
// (nil []string)
// false
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// 2009-02-13 23:31:30 +0000 UTC
// 42
// 0
