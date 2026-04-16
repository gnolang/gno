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
	// ErrNodeNotFound is returned by nodeDB.GetNode when the requested
	// node's record is absent from the DB. This is the only error
	// GetNode surfaces to callers; any other failure (DB read error,
	// deserialization corruption) panics. Callers interpret this as
	// "the node has been pruned or was never persisted". See Finding #5.
	ErrNodeNotFound = errors.New("bptree: node not found")
	// ErrUnsupported is returned by public API methods that are part of
	// the IAVL-compatible surface but have no safe implementation in
	// bptree (they would leak values or nodes). Callers that hit one of
	// these methods must switch to the documented alternative
	// (typically PruneVersionsTo). See Finding #12.
	ErrUnsupported = errors.New("bptree: operation not supported")
)
