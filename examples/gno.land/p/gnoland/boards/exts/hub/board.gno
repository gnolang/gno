package hub

import (
	"gno.land/p/gnoland/boards"
)

// Member defines a type for board members.
type Member struct {
	Address address
	Roles   []string
}

// Board defines a safe type for boards.
type Board struct {
	// id is the unique identifier of the board.
	id uint64

	// name is the current name of the board.
	name string

	// aliases contains a list of alternative names for the board.
	aliases []string

	// readonly indicates that the board is readonly.
	readonly bool

	// threadCount contains the number of threads within the board.
	threadCount int

	// memberCount contains the number of members of the board.
	memberCount int

	// creator is the account address that created the board.
	creator address

	// createdAt is the board's creation time as Unix time.
	createdAt int64

	// updatedAt is the board's update time as Unix time.
	updatedAt int64
}

// ID returns the unique identifier of the board.
func (b Board) ID() uint64 { return b.id }

// Name returns the current name of the board.
func (b Board) Name() string { return b.name }

// Aliases returns the list of alternative names for the board.
func (b Board) Aliases() []string { return b.aliases }

// Readonly indicates that the board is readonly.
func (b Board) Readonly() bool { return b.readonly }

// ThreadCount returns the number of threads within the board.
func (b Board) ThreadCount() int { return b.threadCount }

// MemberCount returns the number of members of the board.
func (b Board) MemberCount() int { return b.memberCount }

// Creator returns the account address that created the board.
func (b Board) Creator() address { return b.creator }

// CreatedAt returns the board's creation time as Unix time.
func (b Board) CreatedAt() int64 { return b.createdAt }

// UpdatedAt returns the board's update time as Unix time.
func (b Board) UpdatedAt() int64 { return b.updatedAt }

// NewSafeBoard creates a safe board.
func NewSafeBoard(ref *boards.Board) Board {
	var usersCount int
	if ref.Permissions != nil {
		usersCount = ref.Permissions.UsersCount()
	}

	var threadCount int
	if ref.Threads != nil {
		threadCount = ref.Threads.Size()
	}

	return Board{
		id:          uint64(ref.ID),
		name:        ref.Name,
		aliases:     append([]string(nil), ref.Aliases...),
		readonly:    ref.Readonly,
		threadCount: threadCount,
		memberCount: usersCount,
		creator:     ref.Creator,
		createdAt:   timeToUnix(ref.CreatedAt),
		updatedAt:   timeToUnix(ref.UpdatedAt),
	}
}
