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
	// ErrNoValueResolver is returned by read APIs that need to resolve
	// a valueKey to raw value bytes when the tree has no resolver
	// configured. Distinguishing this from ErrKeyDoesNotExist lets
	// callers tell a missing key from a misconfigured tree. See
	// Findings #10 and #11.
	ErrNoValueResolver = errors.New("bptree: no value resolver configured")
	// ErrHeightInvariantViolated is returned by prune when a loaded
	// InnerNode claims height == 1 but at least one of its children
	// deserialises as an InnerNode (or vice versa). The leaf-skip
	// optimisation in markReachable / sweepOld would then incorrectly
	// treat that child as a leaf — silently leaking its subtree on
	// prune. Surface this as a typed error rather than a panic so the
	// prune caller can roll back to a consistent checkpoint. See
	// Finding #46.
	ErrHeightInvariantViolated = errors.New("bptree: inner node height invariant violated")
	// ErrValueMissing is returned by nodeDB.GetValue (and the resolver
	// chain that wraps it) when the underlying DB has no record for the
	// requested ValueKey. In a healthy DB this is unreachable — every
	// leaf.valueKeys[i] references a ValueKey that SaveValue persisted —
	// so a return of ErrValueMissing signals out-of-band corruption
	// (e.g. the value file was deleted externally) rather than a Get
	// miss. Surfaced as a typed error so callers can distinguish it
	// from a legitimate empty value (which round-trips as a non-nil
	// zero-length slice with err == nil). See ajnavarro PR #5571
	// review.
	ErrValueMissing = errors.New("bptree: value missing for valueKey")
)
