package boards2

import (
	"std"
	"time"
)

// ReadonlyBoard defines an interface for readonly boards.
type ReadonlyBoard interface {
	// ID returns the unique identifier of the board.
	ID() BoardID

	// Name returns the name of the board.
	Name() string

	// Aliases returns a list with previous names assigned to the board.
	Aliases() []string

	// Creator returns board's creator address.
	Creator() std.Address

	// LastPostID returns the ID assigned to the last created thread.
	LastPostID() uint64

	// CreatedAt returns the time when board was created.
	CreatedAt() time.Time

	// IsReadonly returns true when board shouldn't allow the creation of threads and comments.
	IsReadonly() bool

	// MembersCount returns the number of invited board members.
	MembersCount() int

	// IterateMembers allows reading the list of invited board members.
	IterateMembers(offset, count int, fn UsersIterFn)
}

type readonlyBoard struct {
	board *Board
}

func (b readonlyBoard) ID() BoardID {
	return b.board.id
}

func (b readonlyBoard) Name() string {
	return b.board.name
}

func (b readonlyBoard) Aliases() []string {
	aliases := make([]string, len(b.board.aliases))
	for i, s := range b.board.aliases {
		aliases[i] = s
	}
	return aliases
}

func (b readonlyBoard) Creator() std.Address {
	return b.board.creator
}

func (b readonlyBoard) LastPostID() uint64 {
	return b.board.postsCtr
}

func (b readonlyBoard) CreatedAt() time.Time {
	return b.board.createdAt
}

func (b readonlyBoard) IsReadonly() bool {
	return b.board.readonly
}

func (b readonlyBoard) MembersCount() int {
	return b.board.perms.UsersCount()
}

func (b readonlyBoard) IterateMembers(offset, count int, fn UsersIterFn) {
	b.board.perms.IterateUsers(offset, count, fn)
}
