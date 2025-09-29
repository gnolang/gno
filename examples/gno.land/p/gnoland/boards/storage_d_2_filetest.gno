// PKGPATH: gno.land/r/demo/test
//
// Test removing a stored board with name aliases from the storage.
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
	board = boards.MustNew(1, perms{}, boards.WithName("foo"), boards.EnableAliases())
	err := storage.Save(board)
	if err != nil {
		panic(err)
	}

	board.SetName("bar")
	err = storage.Save(board)
	if err != nil {
		panic(err)
	}
}

func main() {
	b, removed := storage.Remove(board.ID())

	println(removed)
	println(b.ID())

	_, found := storage.GetByName("foo")
	println(found)

	_, found = storage.GetByName("bar")
	println(found)
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
// true
// 1
// false
// false
