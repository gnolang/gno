package boards

import "time"

// Board defines a type for boards.
type Board struct {
	// ID is the unique identifier of the board.
	ID ID

	// Name is the current name of the board.
	Name string

	// Aliases contains a list of alternative names for the board.
	Aliases []string

	// Readonly indicates that the board is readonly.
	Readonly bool

	// Threads contains all board threads.
	Threads PostStorage

	// ThreadsSequence generates sequential ID for new threads.
	ThreadsSequence IdentifierGenerator

	// Permissions enables support for permissioned boards.
	// This type of boards allows managing members with roles and permissions.
	// It also enables the implementation of permissioned execution of board related features.
	Permissions Permissions

	// Creator is the account address that created the board.
	Creator address

	// CreatedAt is the board's creation time.
	CreatedAt time.Time

	// UpdatedAt is the board's update time.
	UpdatedAt time.Time
}

// New creates a new basic non permissioned board.
func New(id ID) *Board {
	return &Board{
		ID:              id,
		Threads:         NewPostStorage(),
		ThreadsSequence: NewIdentifierGenerator(),
		CreatedAt:       time.Now(),
	}
}
