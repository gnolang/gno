// PKGPATH: gno.land/r/demo/test
//
// Test getting a stored board from the storage.
package test

import (
	"std"

	"gno.land/p/gnoland/boards"
)

var (
	board   *boards.Board
	storage = boards.NewStorage()
)

func init() {
	board = boards.MustNew(1, perms{})
	err := storage.Save(board)
	if err != nil {
		panic(err)
	}
}

func main() {
	b, found := storage.Get(board.ID())
	if !found {
		panic("expected board to be found by ID")
	}

	println(b.ID())
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
// 1
