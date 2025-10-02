// PKGPATH: gno.land/r/demo/test
//
// Test reposting a nil thread into another board.
package test

import (
	"gno.land/p/gnoland/boards"
)

var board *boards.Board

func init() {
	board = boards.MustNew(1, perms{})
}

func main() {
	boards.MustNewRepost(nil, board, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
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
// thread to repost is required
