package boards

import (
	"errors"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

const dateFormat = "2006-01-02 3:04pm MST"

type (
	// PostIterFn defines a function type to iterate posts.
	PostIterFn func(*Post) bool

	// MemberIterFn defines a function type to iterate board members.
	MemberIterFn func(address, []Role) bool

	// Board defines a type for boards.
	Board struct {
		id             ID
		name           string
		storage        *Storage
		permissions    Permissions
		lastPostID     seqid.ID
		threads        avl.Tree // string(Post.ID) -> *Post
		aliasesEnabled bool
		aliases        []string
		readonly       bool
		creator        address
		createdAt      time.Time
	}
)

// New creates a new board.
func New(id ID, p Permissions, options ...Option) (*Board, error) {
	if p == nil {
		return nil, errors.New("board permissions is required")
	}

	board := &Board{
		id:          id,
		permissions: p,
		createdAt:   time.Now(),
	}

	for _, apply := range options {
		if err := apply(board); err != nil {
			return nil, err
		}
	}
	return board, nil
}

// MustNew creates a new board or panics on error.
func MustNew(id ID, p Permissions, options ...Option) *Board {
	board, err := New(id, p, options...)
	if err != nil {
		panic(err)
	}
	return board
}

// ID returns the unique identifier of the board.
func (board Board) ID() ID {
	return board.id
}

// Name returns the name of the board.
func (board Board) Name() string {
	return board.name
}

// SetName assigns a new board name.
// It also updates aliases when name aliases is enabled.
func (board *Board) SetName(name string) {
	if board.aliasesEnabled && board.name != "" {
		board.aliases = append([]string{board.name}, board.aliases...)
	}
	board.name = name
}

// Aliases returns the list of names that board had over time.
func (board Board) Aliases() []string {
	return append([]string(nil), board.aliases...)
}

// Readonly returns true when board is readonly.
func (board Board) Readonly() bool {
	return board.readonly
}

// SetReadonly changes board's readonly value.
func (board *Board) SetReadonly(readonly bool) {
	board.readonly = readonly
}

// Creator returns the address of the account that created the board.
func (board Board) Creator() address {
	return board.creator
}

// Permissions returns board permissions.
func (board Board) Permissions() Permissions {
	return board.permissions
}

// CreatedAt returns the time when board was created.
func (board Board) CreatedAt() time.Time {
	return board.createdAt
}

// MembersCount returns the total number of board members.
func (board Board) MembersCount() int {
	return board.permissions.UsersCount()
}

// IterateMembers iterates board members.
func (board Board) IterateMembers(start, count int, fn MemberIterFn) bool {
	return board.permissions.IterateUsers(start, count, func(u User) bool {
		return fn(u.Address, u.Roles)
	})
}

// ThreadsCount returns the total number of board threads.
func (board Board) ThreadsCount() int {
	return board.threads.Size()
}

// IterateThreads iterates board threads.
func (board Board) IterateThreads(start, count int, fn PostIterFn) bool {
	return board.threads.IterateByOffset(start, count, func(_ string, v any) bool {
		p := v.(*Post)
		return fn(p)
	})
}

// ReverseIterateThreads iterates board threads in reverse order.
func (board *Board) ReverseIterateThreads(start, count int, fn PostIterFn) bool {
	return board.threads.ReverseIterateByOffset(start, count, func(_ string, v any) bool {
		p := v.(*Post)
		return fn(p)
	})
}

// GetThread returns a thread from the board.
func (board *Board) GetThread(threadID ID) (_ *Post, found bool) {
	key := makePostKey(threadID)
	v, found := board.threads.Get(key)
	if !found {
		return nil, false
	}
	return v.(*Post), true
}

// DeleteThread deletes a thread from the board.
// NOTE: this can be potentially very expensive for threads with many replies.
func (board *Board) DeleteThread(threadID ID) error {
	// TODO: implement optional fast-delete where thread is simply moved.
	key := makePostKey(threadID)
	if _, removed := board.threads.Remove(key); !removed {
		return errors.New("thread not found, ID=" + threadID.String())
	}
	return nil
}

func (board *Board) generateNextPostID() ID {
	return ID(board.lastPostID.Next())
}

func (board *Board) saveThread(thread *Post) {
	board.threads.Set(makePostKey(thread.ID()), thread)
}
