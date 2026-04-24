package bptree

import "errors"

var (
	ErrVersionDoesNotExist = errors.New("version does not exist")
	ErrKeyDoesNotExist     = errors.New("key does not exist")
	ErrExportDone          = errors.New("export done")
	ErrNotInitializedTree  = errors.New("tree not initialized")
	ErrNoImport            = errors.New("no import in progress")
	ErrNodeMissingNodeKey  = errors.New("node missing node key")
	ErrEmptyTree           = errors.New("tree is empty")
	ErrActiveReaders       = errors.New("version has active readers")
	ErrEmptyKey            = errors.New("key must not be empty")
	ErrKeyTooLong          = errors.New("key exceeds maximum size")
)

// MaxKeyLen caps how long a single key can be. Must stay at or below
// maxReadBytesLen in node.go — if we ever accepted a key larger than the
// read cap, the node would serialize successfully but fail to deserialize,
// permanently wedging that version of the tree.
const MaxKeyLen = 1 << 20 // 1 MiB
