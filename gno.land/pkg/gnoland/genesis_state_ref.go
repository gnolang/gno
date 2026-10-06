package gnoland

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// cacheSchemaVersion identifies the on-disk format of the genesis cache
// produced and consumed by this package. Bump it whenever the layout
// changes (filenames, manifest fields, JSONL element shape, etc.). The
// preprocessor stamps this value into a `version` file in each cache
// directory; the reader rejects any cache whose stamp doesn't match.
const cacheSchemaVersion = 1

// On-disk filenames within a cache directory. The version file holds the
// schema version (and is written last by the preprocessor). The manifest
// holds source hash and counts. The envelope holds all small JSON fields
// (top-level + app_state siblings other than balances/txs). The two JSONL
// files hold one element per line for the bulk arrays.
const (
	versionFilename  = "version"
	manifestFilename = "manifest.json"
	envelopeFilename = "envelope.json"
	balancesFilename = "balances.jsonl"
	txsFilename      = "txs.jsonl"
)

// JSON keys recognized inside the source genesis app_state object. The
// streaming writer recognizes balancesKey and txsKey as bulk arrays; the
// other keys are envelope siblings consumed eagerly by loadAppState.
const (
	appStateAuthKey     = "auth"
	appStateBankKey     = "bank"
	appStateVMKey       = "vm"
	appStateBalancesKey = "balances"
	appStateTxsKey      = "txs"
)

// genesisCacheManifest is the on-disk metadata that describes a preprocessed
// genesis cache directory. Its presence (and matching SourceHash) is the
// signal that the cache is complete and trustworthy.
type genesisCacheManifest struct {
	SourceHash   string `json:"source_hash"`
	BalanceCount int    `json:"balance_count"`
	TxCount      int    `json:"tx_count"`
}

// GenesisStateRef is a lazy, on-disk-backed handle to the bulk fields of a
// genesis app_state. The small sibling fields are loaded into memory eagerly
// (they total under 10 KB in practice); the bulk arrays — balances and txs —
// live as JSONL files and are read one element at a time via the iterators.
//
// Construct via OpenGenesisStateRef. The zero value is not usable.
type GenesisStateRef struct {
	cacheDir    string
	manifest    genesisCacheManifest
	smallFields map[string]json.RawMessage
}

// OpenGenesisStateRef opens an existing genesis cache directory, validates
// it has the expected files, and loads the small sibling fields of app_state
// into memory. The bulk arrays (balances, txs) are not loaded; they are read
// on-demand via IterBalances and IterTxs.
func OpenGenesisStateRef(cacheDir string) (*GenesisStateRef, error) {
	versionBytes, err := os.ReadFile(filepath.Join(cacheDir, versionFilename))
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	gotVersion, err := strconv.Atoi(strings.TrimSpace(string(versionBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse version: %w", err)
	}
	if gotVersion != cacheSchemaVersion {
		return nil, fmt.Errorf("cache schema version mismatch: got %d, want %d", gotVersion, cacheSchemaVersion)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(cacheDir, manifestFilename))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var manifest genesisCacheManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	envelopeBytes, err := os.ReadFile(filepath.Join(cacheDir, envelopeFilename))
	if err != nil {
		return nil, fmt.Errorf("read envelope: %w", err)
	}
	smallFields := map[string]json.RawMessage{}
	if err := json.Unmarshal(envelopeBytes, &smallFields); err != nil {
		return nil, fmt.Errorf("parse envelope: %w", err)
	}

	// balances.jsonl and txs.jsonl are not stat-checked here; they are
	// written before the version file, so a present-and-valid version
	// file implies they exist. Iterators surface a clear error if
	// external corruption removed them.

	return &GenesisStateRef{
		cacheDir:    cacheDir,
		manifest:    manifest,
		smallFields: smallFields,
	}, nil
}

// BalanceCount returns the number of balance entries in the cache, as
// recorded in the manifest. O(1), no file scan.
func (r *GenesisStateRef) BalanceCount() int { return r.manifest.BalanceCount }

// TxCount returns the number of tx entries in the cache, as recorded in the
// manifest. O(1), no file scan.
func (r *GenesisStateRef) TxCount() int { return r.manifest.TxCount }

// SmallField returns the raw JSON bytes of a small sibling field of
// app_state, looked up by key. Returns false if no such field exists.
func (r *GenesisStateRef) SmallField(key string) (json.RawMessage, bool) {
	v, ok := r.smallFields[key]
	return v, ok
}

// IterBalances returns an iterator that yields each balance entry as raw JSON
// bytes (without the trailing newline), one at a time. The iterator stops
// when the file is exhausted, when ctx is cancelled (yielding ctx.Err() once
// before stopping), or when a read error occurs.
//
// The yielded byte slice MUST NOT be retained across iterations — the
// underlying buffer is reused. Copy if you need to keep it.
func (r *GenesisStateRef) IterBalances(ctx context.Context) iter.Seq2[[]byte, error] {
	return r.iterJSONL(ctx, balancesFilename)
}

// IterTxs returns an iterator over the tx entries. Same semantics as
// IterBalances; see that function for caveats on slice retention and
// cancellation.
func (r *GenesisStateRef) IterTxs(ctx context.Context) iter.Seq2[[]byte, error] {
	return r.iterJSONL(ctx, txsFilename)
}

func (r *GenesisStateRef) iterJSONL(ctx context.Context, filename string) iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		if err := ctx.Err(); err != nil {
			yield(nil, err)
			return
		}

		f, err := os.Open(filepath.Join(r.cacheDir, filename))
		if err != nil {
			yield(nil, fmt.Errorf("open %s: %w", filename, err))
			return
		}
		defer f.Close()

		// 1 MB read buffer mirrors the writer side and absorbs the
		// outlier-tx line (~569 KB on the production genesis) without
		// repeated 4 KB refills.
		br := bufio.NewReaderSize(f, 1<<20)
		for {
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}

			line, err := br.ReadBytes('\n')
			if len(line) > 0 {
				// JSONL files may lack a trailing newline on the last line.
				if line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				if !yield(line, nil) {
					return
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					yield(nil, fmt.Errorf("read %s: %w", filename, err))
				}
				return
			}
		}
	}
}

// StreamJSON writes the JSON shape of the original app_state object
// directly to w, reconstituted from the on-disk cache: small sibling
// fields appear first (sorted by key for deterministic output), followed
// by the balances and txs arrays read line-by-line from their JSONL
// caches. The bulk arrays never enter the heap as a unit — each line is
// streamed straight to w.
//
// This satisfies rpctypes.StreamableResult so a *GenesisStateRef can be
// served via the /genesis RPC handler without buffering the whole
// app_state in memory.
//
// ctx is checked between writes; cancellation aborts the stream and
// returns ctx.Err.
func (r *GenesisStateRef) StreamJSON(ctx context.Context, w io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `{`); err != nil {
		return err
	}

	// Sorted small-field keys keep the output byte-stable across runs.
	// Map iteration order in Go is randomized; without sorting, the same
	// ref would emit different bytes across calls and break consumers
	// that key on the wire format (e.g. signing, equality cache).
	keys := make([]string, 0, len(r.smallFields))
	for k := range r.smallFields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return fmt.Errorf("marshal small field key %q: %w", k, err)
		}
		if _, err := w.Write(keyJSON); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `:`); err != nil {
			return err
		}
		if _, err := w.Write(r.smallFields[k]); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `,`); err != nil {
			return err
		}
	}

	if err := r.streamJSONLArray(ctx, w, appStateBalancesKey, r.IterBalances(ctx)); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `,`); err != nil {
		return err
	}
	if err := r.streamJSONLArray(ctx, w, appStateTxsKey, r.IterTxs(ctx)); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `}`); err != nil {
		return err
	}
	return nil
}

// streamJSONLArray writes `"<key>":[<line1>,<line2>,...]` to w, pulling
// elements from the JSONL iterator. ctx is checked between elements so a
// cancelled stream returns promptly.
func (r *GenesisStateRef) streamJSONLArray(ctx context.Context, w io.Writer, key string, seq iter.Seq2[[]byte, error]) error {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("marshal array key %q: %w", key, err)
	}
	if _, err := w.Write(keyJSON); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `:[`); err != nil {
		return err
	}

	first := true
	for line, iterErr := range seq {
		if iterErr != nil {
			return fmt.Errorf("read %s: %w", key, iterErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !first {
			if _, err := io.WriteString(w, `,`); err != nil {
				return err
			}
		}
		if _, err := w.Write(line); err != nil {
			return err
		}
		first = false
	}

	if _, err := io.WriteString(w, `]`); err != nil {
		return err
	}
	return nil
}

// LoadStreamingGenesisDoc parses a genesis JSON file in a single streaming
// pass and returns a *types.GenesisDoc whose AppState is a *GenesisStateRef
// backed by an on-disk cache under cacheRoot. The bulk fields under
// app_state (balances, txs) never enter the heap as a unit; they are
// written line-by-line to JSONL files as the decoder walks the source.
//
// Cache hit (source hash unchanged from a previous run): the cache files
// are reused; the source is still walked once to populate the small
// top-level GenesisDoc fields, but app_state is skipped via Token-level
// step-overs (no decode of the bulk arrays).
//
// Cache miss: the cache directory <cacheRoot>/<sha256(source)>/ is
// populated atomically (via tmp dir + rename), and any other entries in
// cacheRoot are removed (single-cache-on-disk policy).
//
// Memory bound: peak heap is O(largest single element), not O(file size).
// Verified at ~3.7 MB peak on a 200 MB real genesis.
//
// logger receives Warn entries for unknown top-level fields (which are
// silently consumed otherwise — the same tolerance GenesisDocFromJSON
// has). A nil logger falls back to slog.Default.
func LoadStreamingGenesisDoc(genesisPath, cacheRoot string, logger *slog.Logger) (*types.GenesisDoc, error) {
	if logger == nil {
		logger = slog.Default()
	}
	hash, err := hashFile(genesisPath)
	if err != nil {
		return nil, fmt.Errorf("hash source: %w", err)
	}
	finalDir := filepath.Join(cacheRoot, hash)
	ref, openErr := OpenGenesisStateRef(finalDir)

	doc := &types.GenesisDoc{}
	if openErr == nil {
		if err := walkGenesis(genesisPath, doc, nil, logger); err != nil {
			return nil, fmt.Errorf("walk for small fields: %w", err)
		}
	} else {
		if err := writeCache(genesisPath, cacheRoot, finalDir, hash, doc, logger); err != nil {
			return nil, err
		}
		ref, err = OpenGenesisStateRef(finalDir)
		if err != nil {
			return nil, fmt.Errorf("open ref: %w", err)
		}
	}
	doc.AppState = ref
	return doc, nil
}

// writeCache walks the genesis file once, populating doc with small
// top-level fields and writing the cache directory atomically. logger
// receives Warn entries for unknown top-level fields encountered during
// the walk.
func writeCache(genesisPath, cacheRoot, finalDir, hash string, doc *types.GenesisDoc, logger *slog.Logger) error {
	if err := osm.EnsureDir(cacheRoot, 0o755); err != nil {
		return fmt.Errorf("create cache root: %w", err)
	}
	tmpDir, err := os.MkdirTemp(cacheRoot, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create tmp cache dir: %w", err)
	}
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			os.RemoveAll(tmpDir)
		}
	}()

	writer, err := newCacheWriter(tmpDir, hash)
	if err != nil {
		return fmt.Errorf("open cache writer: %w", err)
	}
	defer writer.close()

	if err := walkGenesis(genesisPath, doc, writer, logger); err != nil {
		return fmt.Errorf("walk genesis: %w", err)
	}
	if err := writer.finalize(tmpDir); err != nil {
		return fmt.Errorf("finalize cache: %w", err)
	}

	if err := os.RemoveAll(finalDir); err != nil {
		return fmt.Errorf("remove stale cache: %w", err)
	}
	if err := os.Rename(tmpDir, finalDir); err != nil {
		return fmt.Errorf("install cache: %w", err)
	}
	cleanupTmp = false

	if err := gcCacheRoot(cacheRoot, hash); err != nil {
		return fmt.Errorf("gc cache root: %w", err)
	}
	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := tmhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// walkGenesis streams the genesis JSON file once. Recognized top-level
// small fields (genesis_time, chain_id, consensus_params, validators,
// app_hash) are decoded via amino into doc. The app_state object is
// dispatched to writer if non-nil; if writer is nil, the value is skipped
// at the token level (no allocation of bulk arrays).
//
// Unknown top-level keys are silently consumed (matching GenesisDocFromJSON
// tolerance) but logged via logger.Warn so a misspelled known field
// surfaces in operator logs rather than disappearing.
func walkGenesis(genesisPath string, doc *types.GenesisDoc, writer *cacheWriter, logger *slog.Logger) error {
	f, err := os.Open(genesisPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := expectDelim(dec, '{'); err != nil {
		return fmt.Errorf("genesis root: %w", err)
	}

	for dec.More() {
		key, err := readKey(dec)
		if err != nil {
			return err
		}

		if key == "app_state" {
			if writer != nil {
				if err := writer.consumeAppState(dec); err != nil {
					return fmt.Errorf("app_state: %w", err)
				}
			} else {
				if err := skipValue(dec); err != nil {
					return fmt.Errorf("skip app_state: %w", err)
				}
			}
			continue
		}

		if err := decodeTopLevelField(dec, key, doc, logger); err != nil {
			return fmt.Errorf("top-level %q: %w", key, err)
		}
	}

	return expectDelim(dec, '}')
}

// decodeTopLevelField decodes one recognized top-level field into doc via
// amino (which handles polymorphic crypto.PubKey under validators and
// stringly-encoded int64 for power). Unknown fields are silently consumed
// to match the tolerance of GenesisDocFromJSON; production hardfork
// tooling emits fields like initial_height that GenesisDoc does not model.
// Unknown fields are logged via logger.Warn so a misspelled known field
// (e.g. "consensus_paramz") surfaces in operator logs.
func decodeTopLevelField(dec *json.Decoder, key string, doc *types.GenesisDoc, logger *slog.Logger) error {
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return fmt.Errorf("read raw: %w", err)
	}
	switch key {
	case "genesis_time":
		return amino.UnmarshalJSON(raw, &doc.GenesisTime)
	case "chain_id":
		return amino.UnmarshalJSON(raw, &doc.ChainID)
	case "consensus_params":
		return amino.UnmarshalJSON(raw, &doc.ConsensusParams)
	case "validators":
		return amino.UnmarshalJSON(raw, &doc.Validators)
	case "app_hash":
		return amino.UnmarshalJSON(raw, &doc.AppHash)
	default:
		logger.Warn("ignoring unknown top-level genesis field", "field", key)
		return nil
	}
}

// skipValue consumes the next JSON value (object, array, or scalar)
// without allocating it. Used on the cache-hit path to walk past
// app_state at token level.
func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if _, isDelim := tok.(json.Delim); !isDelim {
		return nil
	}
	depth := 1
	for depth > 0 {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		if d, ok := t.(json.Delim); ok {
			switch d {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
	}
	return nil
}

// cacheWriter holds the open JSONL files and accumulates manifest data
// during a cache-miss walk. Caller must call close after the walk
// completes (success or failure) and finalize on success only.
type cacheWriter struct {
	tmpDir       string
	manifest     genesisCacheManifest
	envelope     map[string]json.RawMessage
	balancesFile *os.File
	balancesBuf  *bufio.Writer
	txsFile      *os.File
	txsBuf       *bufio.Writer
}

func newCacheWriter(tmpDir, sourceHash string) (*cacheWriter, error) {
	balancesFile, err := os.Create(filepath.Join(tmpDir, balancesFilename))
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", balancesFilename, err)
	}
	txsFile, err := os.Create(filepath.Join(tmpDir, txsFilename))
	if err != nil {
		balancesFile.Close()
		return nil, fmt.Errorf("create %s: %w", txsFilename, err)
	}
	return &cacheWriter{
		tmpDir:   tmpDir,
		manifest: genesisCacheManifest{SourceHash: sourceHash},
		envelope: map[string]json.RawMessage{},
		// 1 MB buffer keeps syscall count down on the multi-million-record
		// balances array; default 4 KB would issue ~50K writes on a 200 MB file.
		balancesFile: balancesFile,
		balancesBuf:  bufio.NewWriterSize(balancesFile, 1<<20),
		txsFile:      txsFile,
		txsBuf:       bufio.NewWriterSize(txsFile, 1<<20),
	}, nil
}

// close releases the JSONL file handles. Idempotent: finalize closes the
// files on its success path, so close acts as the cleanup hook for the
// failure path (deferred by writeCache). Errors are intentionally
// discarded here because the writer is being torn down due to an earlier,
// already-reported failure.
func (w *cacheWriter) close() {
	if w.balancesFile != nil {
		w.balancesFile.Close()
		w.balancesFile = nil
	}
	if w.txsFile != nil {
		w.txsFile.Close()
		w.txsFile = nil
	}
}

// consumeAppState walks the value following an "app_state" key, writing
// balances/txs to JSONL and stuffing other siblings into the envelope.
// Tolerates a null app_state.
func (w *cacheWriter) consumeAppState(dec *json.Decoder) error {
	opened, err := expectDelimOrNull(dec, '{')
	if err != nil {
		return err
	}
	if !opened {
		return nil
	}

	for dec.More() {
		key, err := readKey(dec)
		if err != nil {
			return err
		}

		switch key {
		case appStateBalancesKey:
			n, err := streamArrayToJSONL(dec, w.balancesBuf)
			if err != nil {
				return fmt.Errorf("balances: %w", err)
			}
			w.manifest.BalanceCount = n
		case appStateTxsKey:
			n, err := streamArrayToJSONL(dec, w.txsBuf)
			if err != nil {
				return fmt.Errorf("txs: %w", err)
			}
			w.manifest.TxCount = n
		default:
			var raw json.RawMessage
			if err := dec.Decode(&raw); err != nil {
				return fmt.Errorf("decode app_state.%s: %w", key, err)
			}
			w.envelope[key] = raw
		}
	}

	return expectDelim(dec, '}')
}

// finalize closes the JSONL files (flush → sync → close, all error-
// checked) and writes manifest, envelope, and version files into tmpDir.
// The version file is written last so an interrupted finalize leaves a
// cache that OpenGenesisStateRef will reject.
//
// On any error, the underlying files are still released via close so the
// deferred writer.close() in the caller is a defensive no-op rather than
// a double-close. Sync() before Close() forces the kernel to surface
// late write errors (ENOSPC, EIO) that would otherwise be reported only
// at close time on some filesystems — silently dropping those would
// install a truncated cache that looks valid because the version file
// was written last.
func (w *cacheWriter) finalize(tmpDir string) error {
	if err := closeJSONLFile(w.balancesBuf, w.balancesFile, balancesFilename); err != nil {
		w.balancesFile = nil
		w.txsFile.Close()
		w.txsFile = nil
		return err
	}
	w.balancesFile = nil
	if err := closeJSONLFile(w.txsBuf, w.txsFile, txsFilename); err != nil {
		w.txsFile = nil
		return err
	}
	w.txsFile = nil

	if err := writeJSONFile(filepath.Join(tmpDir, manifestFilename), w.manifest); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	if err := writeJSONFile(filepath.Join(tmpDir, envelopeFilename), w.envelope); err != nil {
		return fmt.Errorf("write envelope: %w", err)
	}
	versionContent := []byte(strconv.Itoa(cacheSchemaVersion) + "\n")
	if err := os.WriteFile(filepath.Join(tmpDir, versionFilename), versionContent, 0o644); err != nil {
		return fmt.Errorf("write version: %w", err)
	}
	return nil
}

// closeJSONLFile flushes the buffered writer, syncs the file to disk so
// the kernel surfaces late write errors, then closes the file. Any of
// the three steps failing surfaces an error tagged with the filename.
func closeJSONLFile(buf *bufio.Writer, f *os.File, filename string) error {
	if err := buf.Flush(); err != nil {
		f.Close()
		return fmt.Errorf("flush %s: %w", filename, err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("sync %s: %w", filename, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", filename, err)
	}
	return nil
}

// streamArrayToJSONL decodes a JSON array element-by-element, writing each
// element as a single line to w. The element is re-encoded compactly so
// that pretty-printed source JSON (newlines, indentation) does not bleed
// into the JSONL output and break line-oriented readers. Tolerates null in
// place of the array. Returns the element count.
func streamArrayToJSONL(dec *json.Decoder, w *bufio.Writer) (int, error) {
	opened, err := expectDelimOrNull(dec, '[')
	if err != nil {
		return 0, err
	}
	if !opened {
		return 0, nil
	}

	var compact bytes.Buffer
	count := 0
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return 0, fmt.Errorf("element %d: %w", count, err)
		}
		compact.Reset()
		if err := json.Compact(&compact, raw); err != nil {
			return 0, fmt.Errorf("compact element %d: %w", count, err)
		}
		if _, err := w.Write(compact.Bytes()); err != nil {
			return 0, fmt.Errorf("write element %d: %w", count, err)
		}
		if err := w.WriteByte('\n'); err != nil {
			return 0, fmt.Errorf("write newline %d: %w", count, err)
		}
		count++
	}

	if err := expectDelim(dec, ']'); err != nil {
		return 0, fmt.Errorf("array close: %w", err)
	}
	return count, nil
}

// readKey reads the next token and asserts it is a string (a JSON object
// key in the position the decoder expects one).
func readKey(dec *json.Decoder) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("read key: %w", err)
	}
	key, ok := tok.(string)
	if !ok {
		return "", fmt.Errorf("expected string key, got %T %v", tok, tok)
	}
	return key, nil
}

// expectDelim consumes the next token and verifies it is the given JSON
// delimiter rune ('{', '}', '[', or ']').
func expectDelim(dec *json.Decoder, want rune) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := tok.(json.Delim)
	if !ok || d != json.Delim(want) {
		return fmt.Errorf("expected %q, got %v", want, tok)
	}
	return nil
}

// expectDelimOrNull consumes the next token and accepts either the given
// open delimiter ('{' or '[') or a JSON null. Returns opened=true if the
// caller should iterate the resulting object/array.
func expectDelimOrNull(dec *json.Decoder, openDelim rune) (opened bool, err error) {
	tok, err := dec.Token()
	if err != nil {
		return false, err
	}
	if tok == nil {
		return false, nil
	}
	d, ok := tok.(json.Delim)
	if !ok || d != json.Delim(openDelim) {
		return false, fmt.Errorf("expected %q or null, got %v", openDelim, tok)
	}
	return true, nil
}

func writeJSONFile(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// gcCacheRoot deletes every entry in cacheRoot whose name is not keepHash.
// This enforces the single-cache-on-disk policy: stale hashes from previous
// genesis files and abandoned `.tmp-*` dirs are removed once the new cache
// is installed.
func gcCacheRoot(cacheRoot, keepHash string) error {
	entries, err := os.ReadDir(cacheRoot)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == keepHash {
			continue
		}
		if err := os.RemoveAll(filepath.Join(cacheRoot, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
