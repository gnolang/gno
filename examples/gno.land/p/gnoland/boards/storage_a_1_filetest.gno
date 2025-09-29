// PKGPATH: gno.land/r/demo/test
//
// Test saving board in the storage where board name already exists.
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
	board = boards.MustNew(1, perms{}, boards.WithName("foo"))
	err := storage.Save(board)
	if err != nil {
		panic(err)
	}
}

func main() {
	err := storage.Save(board)
	if err == nil {
		panic("espected storage save to return an error")
	}

	println(err.Error())
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
// duplicated board name: foo
