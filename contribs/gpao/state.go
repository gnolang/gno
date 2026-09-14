package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// stateFileName is the state's name inside --data-dir.
//
// "state", not "height" or "cursor": the height is the only thing a restart
// needs today, and naming the file after it would make the next fact about a
// run either a second file or a lie.
const stateFileName = "state.json"

// state is what one run leaves behind for the next.
type state struct {
	// ChainID is the chain LastVerifiedHeight counts on. Without it a data
	// directory reused against another chain would resume from a height that
	// means nothing there -- and the default --data-dir is under $GNOHOME,
	// which is exactly the kind of path that gets reused.
	//
	// The chain, not the node: a node can move, be replaced, or be one of
	// several serving the same chain, and none of that invalidates a height.
	ChainID string `json:"chain_id"`

	// LastVerifiedHeight is the highest block whose every submitted package
	// reached a verdict -- not the highest block read. A block carrying no
	// MsgAddPackage counts as verified, so this advances on an idle chain too.
	LastVerifiedHeight int64 `json:"last_verified_height"`
}

// stateStore reads and writes the run's state file.
//
// Not safe for concurrent use, and deliberately not made so: the height is
// written by the verifier goroutine alone, which is the same discipline that
// lets `seen`, `overBudget` and `spent` be plain maps and ints. The one write
// from another goroutine is the --start-height override, which happens before
// the verifier is started.
type stateStore struct {
	path    string
	chainID string

	// height is the last recorded cursor, or noCursor when there is none. Kept
	// in memory so the monotonicity guard in setLastVerifiedHeight costs
	// nothing rather than a file read per block.
	height int64
}

// noCursor means no height has been recorded, which is NOT the same as having
// recorded 0: `-start-height 1` legitimately records "verified through 0", and
// reading that back as "nothing recorded" would let the next bare restart jump
// to the tip and silently abandon the replay the operator asked for.
//
// A negative sentinel rather than a second field, because block heights start
// at 1, so no real cursor can collide with it -- and it makes the monotonicity
// guard below a plain comparison instead of a comparison plus a special case.
const noCursor = int64(-1)

// openStateStore prepares dataDir and loads any state already in it.
//
// A missing file is a first run, not an error. Anything else about the file
// that does not make sense IS an error: this daemon exists to avoid repeating
// work, so quietly deciding it has none to resume from is the one failure it
// must not have. Every refusal here is recoverable by the operator with
// --start-height or --data-dir, and says so.
func openStateStore(dataDir, chainID string) (*stateStore, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to prepare data directory %q: %w", dataDir, err)
	}

	s := &stateStore{
		path:    filepath.Join(dataDir, stateFileName),
		chainID: chainID,
		height:  noCursor,
	}

	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read state %q: %w", s.path, err)
	}

	var loaded state
	if err := json.Unmarshal(data, &loaded); err != nil {
		// Writes are atomic, so this is not a torn write -- something else
		// produced it. Naming the remedy because the alternative is a daemon
		// that refuses to boot and does not say what would fix it.
		return nil, fmt.Errorf("failed to decode state %q (delete it to start "+
			"over, or pass --start-height): %w", s.path, err)
	}
	// No allowance for an absent chain id. Every file this daemon writes
	// carries one, so a file without it is a file something else produced --
	// the same class as the undecodable case above, and adopting it for
	// whatever chain happens to be asking is exactly what the guard prevents.
	if loaded.ChainID != chainID {
		return nil, fmt.Errorf("state %q holds a cursor for chain %q, but this "+
			"run is for chain %q; use a different --data-dir, or pass "+
			"--start-height to overwrite it",
			s.path, loaded.ChainID, chainID)
	}

	s.height = loaded.LastVerifiedHeight
	return s, nil
}

// lastVerifiedHeight is the recorded cursor, or noCursor when nothing has been
// recorded yet.
func (s *stateStore) lastVerifiedHeight() int64 {
	return s.height
}

// setLastVerifiedHeight records h as fully verified.
//
// Monotone by construction rather than by argument: the caller is a single
// goroutine draining a FIFO, so h only ever grows, and refusing a rewind here
// means a future second writer cannot silently undo progress.
func (s *stateStore) setLastVerifiedHeight(h int64) error {
	if h <= s.height {
		return nil
	}
	return s.reset(h)
}

// reset records h as fully verified even when that moves the cursor backwards.
func (s *stateStore) reset(h int64) error {
	// Indented, and not for looks: the file is an operator interface --
	// scenarios and humans grep it -- so `"last_verified_height": 42` with the
	// space is part of its shape. json.Marshal would save ~300ns per block and
	// silently turn every such grep into a match on nothing.
	data, err := json.MarshalIndent(state{ChainID: s.chainID, LastVerifiedHeight: h}, "", "  ")
	if err != nil {
		return err
	}
	if err := osm.WriteFileAtomic(s.path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("failed to write state %q: %w", s.path, err)
	}
	s.height = h
	return nil
}
