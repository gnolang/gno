// PKGPATH: gno.land/r/demo/test
//
// Test getting a stored board by name and aliased name from the storage.
package test

import (
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
	b, found := storage.GetByName(board.Name())
	if !found {
		panic("expected board to be found by name")
	}

	println(b.Name())

	b, found = storage.GetByName("foo")
	if !found {
		panic("expected board to be found by alias")
	}

	println(b.Name())
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
// bar
// bar
