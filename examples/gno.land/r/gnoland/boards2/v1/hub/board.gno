package hub

import (
	"time"

	"gno.land/p/gnoland/boards"
)

// Member defines a type for board members.
type Member struct {
	Address address
	Roles   []string
}

// Board defines a safe type for boards.
type Board struct {
	ref *boards.Board

	// ID is the unique identifier of the board.
	ID uint64

	// Name is the current name of the board.
	Name string

	// Aliases contains a list of alternative names for the board.
	Aliases []string

	// Readonly indicates that the board is readonly.
	Readonly bool

	// ThreadsCount contains the number of threads within the board.
	ThreadCount int

	// MemberCount contains the number of members of the board.
	MemberCount int

	// Creator is the account address that created the board.
	Creator address

	// CreatedAt is the board's creation time.
	CreatedAt time.Time

	// UpdatedAt is the board's update time.
	UpdatedAt time.Time
}

// IterateThreads iterates board threads by creation time.
// To reverse iterate use a negative count.
func (b Board) IterateThreads(start, count int, fn func(Thread) bool) bool {
	return b.ref.Threads.Iterate(start, count, func(thread *boards.Post) bool {
		return fn(newSafeThread(thread))
	})
}

// IterateMembers iterates board members.
// To reverse iterate use a negative count.
func (b Board) IterateMembers(start, count int, fn func(boards.User) bool) bool {
	return b.ref.Permissions.IterateUsers(start, count, fn)
}

func newSafeBoard(ref *boards.Board) Board {
	return Board{
		ref:         ref,
		ID:          uint64(ref.ID),
		Name:        ref.Name,
		Aliases:     append([]string(nil), ref.Aliases...),
		Readonly:    ref.Readonly,
		ThreadCount: ref.Threads.Size(),
		MemberCount: ref.Permissions.UsersCount(),
		Creator:     ref.Creator,
		CreatedAt:   ref.CreatedAt,
		UpdatedAt:   ref.UpdatedAt,
	}
}
