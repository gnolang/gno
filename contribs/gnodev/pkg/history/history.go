// Package history persists a gnodev chain's transaction history to a directory
// on disk, so that a run's output is the next run's input.
//
// Segments hold one amino-JSON gnoland.TxWithMetadata per line. That is the
// same encoding contribs/tx-archive writes, that gnodev's own -txs-file and
// gnoland start --genesis-txs-file read, that gnogenesis fork generate takes as
// --source-txs-jsonl-file, and that gnolang/tx-exports stores. A segment can
// therefore be committed to an archive or replayed by any of those tools with
// no conversion step.
package history

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// DefaultMaxSegmentBytes caps how large one segment grows before the next
// transaction opens a new one. Segments are append-only, so the cap exists to
// keep any single file diffable, uploadable and cheap to re-read, not to bound
// total history.
const DefaultMaxSegmentBytes = 64 << 20 // 64 MiB

// maxSegmentLineBytes caps a single line. One line is one transaction, and an
// addpkg transaction carries whole package sources, so the ceiling has to be
// generous rather than bufio's 64KiB default.
const maxSegmentLineBytes = 64 << 20 // 64 MiB

// Config configures a Store. Only ChainID is load-bearing; the rest have
// working defaults.
type Config struct {
	// ChainID is recorded in the manifest and checked against it on open, so
	// a state dir cannot silently be reused for a different chain.
	ChainID string
	// ChainDomain is recorded for the benefit of whoever reads the manifest.
	ChainDomain string
	// MaxSegmentBytes overrides DefaultMaxSegmentBytes when positive.
	MaxSegmentBytes int64
	// Logger receives the one-line summaries this package emits. Optional.
	Logger *slog.Logger
}

// Store is an append-only transaction log under a single directory.
//
// Record buffers, so it is cheap enough to call from the node's block
// execution path; durability is established by Sync, which the caller drives
// (gnodev syncs on a ticker, before every node reload, and on shutdown).
type Store struct {
	dir     string
	histDir string
	cfg     Config
	logger  *slog.Logger

	mu       sync.Mutex
	manifest *Manifest
	segment  *os.File
	writer   *bufio.Writer
	segBytes int64
	segIndex int
	recorded int // transactions appended by this process
	loaded   int // transactions read back by Load
	closed   bool
}

// Open prepares dir as a state directory, creating it when absent, and starts
// a fresh segment for this process.
//
// It fails rather than guesses when the directory already belongs to another
// chain or to a newer schema: appending one chain's transactions to another's
// history produces a replay that cannot be reasoned about.
func Open(dir string, cfg Config) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("no state directory given")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.MaxSegmentBytes <= 0 {
		cfg.MaxSegmentBytes = DefaultMaxSegmentBytes
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve state dir %q: %w", dir, err)
	}
	histDir := filepath.Join(abs, historyDirname)
	if err := os.MkdirAll(histDir, 0o755); err != nil {
		return nil, fmt.Errorf("create state dir %q: %w", histDir, err)
	}

	manifest, err := readManifest(abs)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if manifest == nil {
		manifest = &Manifest{
			SchemaVersion: SchemaVersion,
			ChainID:       cfg.ChainID,
			ChainDomain:   cfg.ChainDomain,
			CreatedAt:     now,
		}
	} else if err := manifest.checkCompatible(cfg.ChainID); err != nil {
		return nil, err
	}
	// Adopt the current identity: an older manifest may predate these fields.
	manifest.SchemaVersion = SchemaVersion
	manifest.ChainID = cfg.ChainID
	manifest.ChainDomain = cfg.ChainDomain

	s := &Store{
		dir:      abs,
		histDir:  histDir,
		cfg:      cfg,
		logger:   cfg.Logger,
		manifest: manifest,
	}

	existing, err := s.segmentPaths()
	if err != nil {
		return nil, err
	}
	s.segIndex = len(existing)

	return s, nil
}

// Dir returns the resolved state directory.
func (s *Store) Dir() string { return s.dir }

// segmentPaths lists existing segments in replay order.
//
// Ordering is numeric on the filename stem rather than lexicographic, so a
// history that outgrows six digits still replays in the order it was written.
func (s *Store) segmentPaths() ([]string, error) {
	entries, err := os.ReadDir(s.histDir)
	if err != nil {
		return nil, fmt.Errorf("read history dir: %w", err)
	}

	type seg struct {
		n    int
		path string
	}
	var segs []seg
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), segmentExt) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSuffix(e.Name(), segmentExt))
		if err != nil {
			// Not one of ours. Leave it alone and say so once, rather than
			// replaying a file whose ordering we cannot establish.
			s.logger.Warn("ignoring unrecognized file in history dir", "name", e.Name())
			continue
		}
		segs = append(segs, seg{n: n, path: filepath.Join(s.histDir, e.Name())})
	}
	sort.Slice(segs, func(i, j int) bool { return segs[i].n < segs[j].n })

	paths := make([]string, len(segs))
	for i, sg := range segs {
		paths[i] = sg.path
	}
	return paths, nil
}

// Load reads every existing segment in order. The result is suitable as
// gnodev's genesis transaction list.
//
// A truncated trailing line is tolerated with a warning: it is what a kill
// mid-flush leaves behind, and dropping the partial transaction is both the
// only recoverable choice and a strictly smaller loss than refusing to boot.
func (s *Store) Load(ctx context.Context) ([]gnoland.TxWithMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	paths, err := s.segmentPaths()
	if err != nil {
		return nil, err
	}

	var txs []gnoland.TxWithMetadata
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		loaded, err := readSegment(ctx, path, s.logger)
		if err != nil {
			return nil, err
		}
		txs = append(txs, loaded...)
	}

	s.loaded = len(txs)
	if len(txs) > 0 {
		s.logger.Info("history loaded", "txs", len(txs), "segments", len(paths), "dir", s.dir)
	} else {
		s.logger.Info("history is empty, starting a new chain", "dir", s.dir)
	}
	return txs, nil
}

func readSegment(ctx context.Context, path string, logger *slog.Logger) ([]gnoland.TxWithMetadata, error) {
	// A segment written right up to a hard kill can end mid-line. That is
	// recoverable, and only for the final line; a bad line anywhere else is
	// corruption we must not paper over.
	complete, err := endsWithNewline(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open segment %q: %w", path, err)
	}
	defer f.Close()

	var (
		txs     []gnoland.TxWithMetadata
		scanner = bufio.NewScanner(f)
	)
	// Genesis addpkg transactions carry whole package sources, so a single
	// line can be far larger than bufio's 64KiB default.
	scanner.Buffer(make([]byte, 0, 1<<20), maxSegmentLineBytes)

	for line := 1; scanner.Scan(); line++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		var tx gnoland.TxWithMetadata
		if err := amino.UnmarshalJSON([]byte(raw), &tx); err != nil {
			// Tolerate only a partial final line in a segment with no
			// terminating newline. Consuming one more token is how we tell
			// "last line" from "line in the middle"; we are returning either
			// way, so the consumed token does not matter.
			if !complete && !scanner.Scan() {
				logger.Warn("history segment ends mid-line, dropping the partial transaction",
					"segment", path, "line", line)
				break
			}
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		txs = append(txs, tx)
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("%s: a line exceeds the %d byte limit: %w", path, maxSegmentLineBytes, err)
		}
		return nil, fmt.Errorf("read segment %q: %w", path, err)
	}
	return txs, nil
}

// endsWithNewline reports whether the file is newline-terminated, which is
// what distinguishes a fully written segment from one cut short. An empty file
// counts as complete: it has no partial line.
func endsWithNewline(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("open segment %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("stat segment %q: %w", path, err)
	}
	if info.Size() == 0 {
		return true, nil
	}

	last := make([]byte, 1)
	if _, err := f.ReadAt(last, info.Size()-1); err != nil {
		return false, fmt.Errorf("read tail of segment %q: %w", path, err)
	}
	return last[0] == '\n', nil
}

// Record appends one transaction. It buffers, and does not fsync; call Sync to
// make previous Records durable.
func (s *Store) Record(tx gnoland.TxWithMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("history store is closed")
	}

	raw, err := encodeLine(tx)
	if err != nil {
		return err
	}

	if s.writer == nil || s.segBytes+int64(len(raw)) > s.cfg.MaxSegmentBytes {
		if err := s.rotateLocked(); err != nil {
			return err
		}
	}

	n, err := s.writer.Write(raw)
	s.segBytes += int64(n)
	if err != nil {
		return fmt.Errorf("append transaction: %w", err)
	}
	s.recorded++
	return nil
}

// encodeLine renders one transaction as a segment line: amino JSON plus the
// terminating newline.
func encodeLine(tx gnoland.TxWithMetadata) ([]byte, error) {
	raw, err := amino.MarshalJSON(tx)
	if err != nil {
		return nil, fmt.Errorf("encode transaction: %w", err)
	}
	return append(raw, '\n'), nil
}

// rotateLocked closes the open segment, if any, and opens the next one.
func (s *Store) rotateLocked() error {
	if err := s.closeSegmentLocked(); err != nil {
		return err
	}

	s.segIndex++
	path := filepath.Join(s.histDir, fmt.Sprintf("%06d%s", s.segIndex, segmentExt))

	// O_EXCL: a collision means the index was computed from a stale listing,
	// and appending to a segment another process holds open would interleave
	// two chains' transactions in one file.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("open segment %q: %w", path, err)
	}

	s.segment = f
	s.writer = bufio.NewWriterSize(f, 1<<16)
	s.segBytes = 0

	s.manifest.Epochs++
	s.logger.Info("history segment opened", "segment", filepath.Base(path))
	return nil
}

func (s *Store) closeSegmentLocked() error {
	if s.segment == nil {
		return nil
	}
	if err := s.writer.Flush(); err != nil {
		s.segment.Close()
		s.segment, s.writer = nil, nil
		return fmt.Errorf("flush segment: %w", err)
	}
	if err := s.segment.Sync(); err != nil {
		s.segment.Close()
		s.segment, s.writer = nil, nil
		return fmt.Errorf("sync segment: %w", err)
	}
	err := s.segment.Close()
	s.segment, s.writer = nil, nil
	if err != nil {
		return fmt.Errorf("close segment: %w", err)
	}
	return nil
}

// Sync flushes buffered transactions to the filesystem and refreshes the
// manifest. Everything Recorded before it returns survives a crash.
func (s *Store) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.syncLocked()
}

func (s *Store) syncLocked() error {
	if s.writer != nil {
		if err := s.writer.Flush(); err != nil {
			return fmt.Errorf("flush segment: %w", err)
		}
		if err := s.segment.Sync(); err != nil {
			return fmt.Errorf("sync segment: %w", err)
		}
	}

	s.manifest.Txs = s.loaded + s.recorded
	s.manifest.UpdatedAt = time.Now().UTC()
	return writeManifest(s.dir, s.manifest)
}

// Close flushes and releases the open segment. It is safe to call twice.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	// Sync first: closeSegmentLocked drops the handle syncLocked needs.
	syncErr := s.syncLocked()
	closeErr := s.closeSegmentLocked()

	s.logger.Info("history closed", "recorded", s.recorded, "total", s.manifest.Txs, "dir", s.dir)

	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

// Recorded returns how many transactions this process has appended.
func (s *Store) Recorded() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recorded
}

// Total returns every transaction in this directory: those read back at boot
// plus those appended since.
func (s *Store) Total() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loaded + s.recorded
}

var _ io.Closer = (*Store)(nil)
