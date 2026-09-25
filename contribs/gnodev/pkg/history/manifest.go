package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SchemaVersion identifies the on-disk layout of a state directory. Bump it
// whenever the layout changes (filenames, manifest fields, segment encoding),
// and Open will refuse a directory written by a newer gnodev rather than
// appending to something it does not understand.
const SchemaVersion = 1

const (
	manifestFilename = "MANIFEST.json"
	historyDirname   = "history"
	segmentExt       = ".jsonl"
)

// Manifest is the small bookkeeping file at the root of a state directory. It
// exists for two reasons: to refuse a directory that belongs to a different
// chain, and to let a human (or a script about to push the history to an
// archive) see what is in there without parsing every segment.
type Manifest struct {
	SchemaVersion int    `json:"schema_version"`
	ChainID       string `json:"chain_id"`
	ChainDomain   string `json:"chain_domain,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Epochs counts how many times a gnodev process has opened this
	// directory. Each epoch opens a fresh segment, so this is also the
	// number of segments ever created.
	Epochs int `json:"epochs"`
	// Txs is the number of transactions across every segment, as of
	// UpdatedAt. Advisory: the segments are the source of truth, and a
	// crash can leave this trailing the files by one flush interval.
	Txs int `json:"txs"`
}

func manifestPath(dir string) string { return filepath.Join(dir, manifestFilename) }

// readManifest loads the manifest, returning a nil manifest and no error when
// the directory holds none yet (a fresh state dir).
func readManifest(dir string) (*Manifest, error) {
	raw, err := os.ReadFile(manifestPath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", manifestFilename, err)
	}
	return &m, nil
}

// writeManifest replaces the manifest atomically, so a crash mid-write leaves
// the previous one intact rather than a truncated file that fails to parse and
// makes the whole directory unopenable.
func writeManifest(dir string, m *Manifest) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	raw = append(raw, '\n')

	tmp, err := os.CreateTemp(dir, manifestFilename+".*")
	if err != nil {
		return fmt.Errorf("create temp manifest: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp manifest: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp manifest: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp manifest: %w", err)
	}
	// CreateTemp makes 0600 files and Rename keeps the mode, which would
	// leave the manifest unreadable to anyone but the writing user: wrong for
	// a directory meant to be inspected and archived.
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("chmod temp manifest: %w", err)
	}
	if err := os.Rename(tmpName, manifestPath(dir)); err != nil {
		return fmt.Errorf("install manifest: %w", err)
	}
	return nil
}

// checkCompatible reports why an existing manifest cannot be reused, if so.
func (m *Manifest) checkCompatible(chainID string) error {
	if m.SchemaVersion > SchemaVersion {
		return fmt.Errorf("state dir has schema version %d, this gnodev understands up to %d: upgrade gnodev or point -state-dir elsewhere",
			m.SchemaVersion, SchemaVersion)
	}
	if m.ChainID != "" && chainID != "" && m.ChainID != chainID {
		return fmt.Errorf("state dir holds history for chain %q but -chain-id is %q: replaying one chain's transactions onto another is never what you want",
			m.ChainID, chainID)
	}
	return nil
}
