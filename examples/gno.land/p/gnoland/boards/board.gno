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
	// TODO: Keep hidden threads in a different storage
	Threads PostStorage

	// ThreadsSequence generates sequential ID for new threads.
	ThreadsSequence IdentifierGenerator

	// Permissions enables support for permissioned boards.
	// This type of boards allows managing members with roles and permissions.
	// It also enables the implementation of permissioned execution of board related features.
	Permissions Permissions

	// Creator is the account address that created the board.
	Creator address

	// Meta allows storing board metadata.
	Meta any

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

// SetID sets board ID value.
func (board *Board) SetID(v ID) {
	board.ID = v
}

// SetName sets name value.
func (board *Board) SetName(v string) {
	board.Name = v
}

// SetAliases sets board name aliases.
func (board *Board) SetAliases(v []string) {
	board.Aliases = v
}

// SetReadonly sets readonly value.
func (board *Board) SetReadonly(v bool) {
	board.Readonly = v
}

// SetThreadStorage sets the storage where board threads are stored.
func (board *Board) SetThreadStorage(v PostStorage) {
	board.Threads = v
}

// SetThreadsSequence sets the sequential thread ID generator.
func (board *Board) SetThreadsSequence(v IdentifierGenerator) {
	board.ThreadsSequence = v
}

// SetPermissions sets permissions value.
func (board *Board) SetPermissions(v Permissions) {
	board.Permissions = v
}

// SetCreator sets the address of the account that created the board.
func (board *Board) SetCreator(v address) {
	board.Creator = v
}

// SetCreatedAt sets the time when board was created.
func (board *Board) SetCreatedAt(v time.Time) {
	board.CreatedAt = v
}

// SetUpdatedAt sets the time when a board value was updated.
func (board *Board) SetUpdatedAt(v time.Time) {
	board.UpdatedAt = v
}

// SetMeta sets board metadata.
func (board *Board) SetMeta(v any) {
	board.Meta = v
}
