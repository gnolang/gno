package wal

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	goerrors "errors"
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	auto "github.com/gnolang/gno/tm2/pkg/autofile"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/service"
)

const (
	// how often the WAL should be sync'd during period sync'ing
	walDefaultFlushInterval = 2 * time.Second
)

var (
	crc32c      = crc32.MakeTable(crc32.Castagnoli)
	base64stdnp = base64.StdEncoding.WithPadding(base64.NoPadding)
)

// --------------------------------------------------------
// types and functions for savings consensus messages

type WALMessage interface {
	AssertWALMessage()
}

// TimedWALMessage wraps WALMessage and adds Time for debugging purposes.
type TimedWALMessage struct {
	Time time.Time  `json:"time"`
	Msg  WALMessage `json:"msg"`
}

// Some lines are MetaMessages to denote new heights, etc.
// NOTE: The encoding must not contain the '#' character,
// so arbitrary strings will not work without some kind of escaping.
// TODO: consider alternative meta line schemas in the long run, or escape the
// '#' character in such a way that scanning randomly from any position in a
// file resumes correctly after the first and only likely corruption.
type MetaMessage struct {
	Height int64 `json:"h"`
}

// --------------------------------------------------------
// Simple write-ahead logger

// WAL is an interface for any write-ahead logger.
type WAL interface {
	// config methods
	SetLogger(l *slog.Logger)

	// write methods
	Write(WALMessage) error
	WriteSync(WALMessage) error
	WriteMetaSync(MetaMessage) error
	FlushAndSync() error

	// search methods
	SearchForHeight(height int64, options *WALSearchOptions) (rd io.ReadCloser, found bool, err error)

	// service methods
	Start() error
	Stop() error
	Wait()
}

// Write ahead logger writes msgs to disk before they are processed.
// Can be used for crash-recovery and deterministic replay.
// TODO: currently the wal is overwritten during replay catchup, give it a mode
// so it's either reading or appending - must read to end to start appending
// again.
type baseWAL struct {
	service.BaseService

	group *auto.Group

	maxSize int64
	enc     *WALWriter

	flushTicker   *time.Ticker
	flushInterval time.Duration
}

var _ WAL = &baseWAL{}

// NewWAL returns a new write-ahead logger based on `baseWAL`, which implements
// WAL. It's flushed and synced to disk every 2s and once when stopped.
// `maxSize` is the maximum allowable amino bytes of a TimedWALMessage
// including the amino (byte) size prefix, but excluding any crc checks.
func NewWAL(walFile string, maxSize int64, groupOptions ...func(*auto.Group)) (*baseWAL, error) {
	err := osm.EnsureDir(filepath.Dir(walFile), 0o700)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure WAL directory is in place")
	}

	group, err := auto.OpenGroup(walFile, groupOptions...)
	if err != nil {
		return nil, err
	}
	wal := &baseWAL{
		group:         group,
		maxSize:       maxSize,
		enc:           NewWALWriter(group, maxSize),
		flushInterval: walDefaultFlushInterval,
	}
	wal.BaseService = *service.NewBaseService(nil, "baseWAL", wal)
	return wal, nil
}

// SetFlushInterval allows us to override the periodic flush interval for the WAL.
func (wal *baseWAL) SetFlushInterval(i time.Duration) {
	wal.flushInterval = i
}

func (wal *baseWAL) Group() *auto.Group {
	return wal.group
}

func (wal *baseWAL) SetLogger(l *slog.Logger) {
	wal.BaseService.Logger = l
	wal.group.SetLogger(l)
}

func (wal *baseWAL) OnStart() error {
	size := wal.group.ReadGroupInfo().TotalSize
	if size == 0 {
		wal.WriteMetaSync(MetaMessage{Height: 0})
	}
	err := wal.group.Start()
	if err != nil {
		return err
	}
	wal.flushTicker = time.NewTicker(wal.flushInterval)
	go wal.processFlushTicks()
	return nil
}

func (wal *baseWAL) processFlushTicks() {
	for {
		select {
		case <-wal.flushTicker.C:
			if err := wal.FlushAndSync(); err != nil {
				wal.Logger.Error("Periodic WAL flush failed", "err", err)
			}
		case <-wal.Quit():
			return
		}
	}
}

// FlushAndSync flushes and fsync's the underlying group's data to disk.
// See auto#FlushAndSync
func (wal *baseWAL) FlushAndSync() error {
	return wal.group.FlushAndSync()
}

// Stop the underlying autofile group.
// Use Wait() to ensure it's finished shutting down
// before cleaning up files.
func (wal *baseWAL) OnStop() {
	wal.flushTicker.Stop()
	wal.FlushAndSync()
	wal.group.Stop()
	wal.group.Close()
}

// Wait for the underlying autofile group to finish shutting down
// so it's safe to cleanup files.
func (wal *baseWAL) Wait() {
	wal.group.Wait()
}

// Write is called in newStep and for each receive on the
// peerMsgQueue and the timeoutTicker.
// NOTE: does not call fsync()
func (wal *baseWAL) Write(msg WALMessage) error {
	if wal == nil {
		return nil
	}

	if err := wal.enc.Write(TimedWALMessage{tmtime.Now(), msg}); err != nil {
		wal.Logger.Error("Error writing msg to consensus wal. WARNING: recover may not be possible for the current height",
			"err", err, "msg", msg)
		return err
	}

	return nil
}

// WriteMetaSync writes the new height and finalizes the previous height.
// NOTE: It is useless to implement WriteMeta() (asynchronous) because there is
// usually something to do in sync after the aforementioned finalization
// occurs.
func (wal *baseWAL) WriteMetaSync(meta MetaMessage) error {
	if wal == nil {
		return nil
	}

	if err := wal.enc.WriteMeta(meta); err != nil {
		wal.Logger.Error("Error writing height to consensus wal. WARNING: full recover may not be possible for the previous height",
			"err", err)
		return err
	}

	if err := wal.FlushAndSync(); err != nil {
		wal.Logger.Error("WriteSync failed to flush consensus wal. WARNING: may result in creating alternative proposals / votes for the current height iff the node restarted",
			"err", err)
		return err
	}

	return nil
}

// WriteSync is called when we receive a msg from ourselves
// so that we write to disk before sending signed messages.
// NOTE: calls fsync()
func (wal *baseWAL) WriteSync(msg WALMessage) error {
	if wal == nil {
		return nil
	}

	if err := wal.Write(msg); err != nil {
		return err
	}

	if err := wal.FlushAndSync(); err != nil {
		wal.Logger.Error("WriteSync failed to flush consensus wal. WARNING: may result in creating alternative proposals / votes for the current height iff the node restarted",
			"err", err)
		return err
	}

	return nil
}

type WALSearchMode int

const (
	WALSearchModeInvalid WALSearchMode = iota
	WALSearchModeBackwards
	WALSearchModeBinary
)

// WALSearchOptions are optional arguments to SearchForHeight.
type WALSearchOptions struct {
	// Mode is by default backwards via WALSearchModeBackwards (limit at 100
	// blocks).
	Mode WALSearchMode
	// IgnoreDataCorruptionErrors set to true will result in skipping data
	// corruption errors (default false).
	IgnoreDataCorruptionErrors bool
}

// SearchForHeight scans meta lines to find the first line after height as
// denoted by a meta line, and returns an auto.GroupReader.  Group reader will
// be nil if found equals false.
//
// The result is a buffered ReadCloser.
//
// CONTRACT: caller must close group reader.
func (wal *baseWAL) SearchForHeight(height int64, options *WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
	var (
		meta *MetaMessage
		gr   *auto.GroupReader
	)

	// NOTE: starting from the last file in the group because we're usually
	// searching for the last height. See replay.go
	minVal, maxVal := wal.group.MinIndex(), wal.group.MaxIndex()
	wal.Logger.Info("Searching for height", "height", height, "min", minVal, "max", maxVal)

	var (
		mode    = WALSearchModeBackwards
		backoff = 0 // backwards search offset (goes 0,-1,-2,-4,-8,-16...)
		idxoff  = 0 // forward index offset (for when file contains no meta lines).
	)

	// option overrides.
	if options != nil && options.Mode != WALSearchModeInvalid {
		mode = options.Mode
	}

OUTER_LOOP:
	for minVal <= maxVal {
		var index int

		// set index depending on mode.
		switch mode {
		case WALSearchModeBackwards:
			index = maxVal + backoff + idxoff
			if maxVal < index {
				// (max+backoff)+ doesn't contain any height.
				// adjust max & backoff accordingly.
				idxoff = 0
				maxVal = maxVal + backoff - 1
				if backoff == 0 {
					backoff = -1
				} else {
					backoff = backoff * 2 // exponential backwards
				}
				continue OUTER_LOOP
			}
			if index < minVal {
				panic("should not happen")
			}
		case WALSearchModeBinary:
			index = (minVal+maxVal+1)/2 + idxoff
			if maxVal < index {
				// ((min+max+1)/2)+ doesn't contain any height.
				// adjust max & binary search accordingly.
				idxoff = 0
				maxVal = (minVal+maxVal+1)/2 - 1
				continue OUTER_LOOP
			}
		}

		gr, err = wal.group.NewReader(index, index+1) // read only for index.
		if err != nil {
			return nil, false, err
		}
		dec := NewWALReader(gr, wal.maxSize)

	FILE_LOOP:
		for {
			_, meta, err = dec.ReadMessage()
			// error case
			if err != nil {
				if goerrors.Is(err, io.EOF) {
					// adjust next index.
					// @index didn't have a height declaration.
					idxoff++
					dec.Close()
					continue OUTER_LOOP
				} else if options != nil && options.IgnoreDataCorruptionErrors && IsDataCorruptionError(err) {
					wal.Logger.Error("Corrupted entry. Skipping...", "err", err)
					continue FILE_LOOP // skip corrupted line and ignore error.
				} else {
					dec.Close()
					return nil, false, err
				}
			}
			// meta case
			if meta != nil {
				if height < meta.Height {
					// adjust next index.
					// it's in an earlier position or file.
					switch mode {
					case WALSearchModeBackwards:
						idxoff = 0
						if backoff == 0 {
							maxVal--
							backoff = -1
						} else {
							maxVal += backoff
							backoff *= 2
						}
						// convert to binary search if backoff is too big.
						// max+backoff would work but max+(backoff*2) is smoother.
						if maxVal+(backoff*2) <= minVal {
							wal.Logger.Info("Converting to binary search",
								"height", height, "min", minVal,
								"max", maxVal, "backoff", backoff)
							backoff = 0
							mode = WALSearchModeBinary
						}
					case WALSearchModeBinary:
						idxoff = 0
						maxVal = (minVal+maxVal+1)/2 - 1
					}
					dec.Close()
					continue OUTER_LOOP
				} else if meta.Height == height { // found
					wal.Logger.Info("Found", "height", height, "index", index)
					// NOTE: dec itself is an io.ReadCloser for this purpose.
					// NOTE: the result is buffered, specifically a bufio.Reader.
					return dec, true, nil
				} else { // meta.Height < height
					// it comes later, maybe @index.
					switch mode {
					case WALSearchModeBackwards:
						if backoff == 0 {
							// ignore and keep reading
							// NOTE: in the future we could binary search
							// within a file, but for now we read sequentially.
							continue FILE_LOOP
						} else {
							// convert to binary search with index as new min.
							wal.Logger.Info("Converting to binary search with new min",
								"height", height, "min", minVal,
								"max", maxVal, "backoff", backoff)
							idxoff = 0
							backoff = 0
							minVal = index
							mode = WALSearchModeBinary
							dec.Close()
							continue OUTER_LOOP
						}
					case WALSearchModeBinary:
						if index < maxVal {
							// maybe in @index, but first try binary search
							// between @index and max.
							idxoff = 0
							minVal = index
							dec.Close()
							continue OUTER_LOOP
						} else { // index == max
							// this is the last file, keep reading.
							// NOTE: in the future we could binary search
							// within a file, but for now we read sequentially.
							continue FILE_LOOP
						}
					}
				}
			}
			// msg case
			// if msg != nil {
			// do nothing, we're only interested in meta lines.
			// TODO: optimize by implementing ReadNextMeta(),
			// which skips decoding non-meta messages.
			// }
		}
	}

	return nil, false, nil
}

// -----------

// A WALWriter writes custom-encoded WAL messages to an output stream.
// Each binary WAL entry is length encoded, then crc encoded,
// then base64 encoded and delimited by newlines.
// The base64 encoding used is encodeStd,
// `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/`.
//
// Each WAL item is also newline delimited.
//
// Meta lines are prefixed with a '#' (which is not a valid base64 character)
// denote meta information, such as new height.  The purpose of base64
// encoding is to enable backwards traversal of items (e.g. in search of some
// previous height).
//
// Base64 is chosen to optimize the worst-case scenario for encoding delimited
// binary bytes while enabling metadata insertion and also supporting
// backwards traversal of binary blobs (to enable binary search, etc). In the
// future, base64 could be replaced with a similar encoding formula, and the
// crc function could change too, but otherwise the structure of the WAL
// should not change, including the rule that all metalines should start with
// '#' and that all items be delimited with a newline. This implementation
// enforces ASCII text characters, but other implementations may choose
// otherwise.
//
// Format: base64(4 bytes CRC sum + length-prefixed amino bytes) + newline.
// e.g.
// ```
// ABCDEFGHIJKLMNOPQRSTUV00
// ABCDEFGHIJKLMNOPQRSTUV01
// ABCDEFGHIJKLMNOPQRSTUV02
// #{"h":"123"}
// ABCDEFGHIJKLMNOPQRSTUV03
// ABCDEFGHIJKLMNOPQRSTUV04
type WALWriter struct {
	wr      io.Writer
	maxSize int64 // max WALMessage amino size excluding time/crc/base64.
}

// NewWALWriter returns a new encoder that writes to wr.
func NewWALWriter(wr io.Writer, maxSize int64) *WALWriter {
	return &WALWriter{wr, maxSize}
}

// Write writes the custom encoding of v to the stream, followed by a newline
// byte. It returns an error if the amino-encoded size of v is greater than
// maxSize. Any error encountered during the write is also returned.
func (enc *WALWriter) Write(v TimedWALMessage) error {
	twmBytes := amino.MustMarshalSized(v)

	length := int64(len(twmBytes))
	if 0 < enc.maxSize && enc.maxSize < length {
		return fmt.Errorf("msg is too big: %d bytes, max: %d bytes", length, enc.maxSize)
	}

	totalLength := 4 + int(length)
	crc := crc32.Checksum(twmBytes, crc32c)
	line := make([]byte, totalLength)
	binary.BigEndian.PutUint32(line[0:4], crc)
	copy(line[4:], twmBytes)

	line64 := base64stdnp.EncodeToString(line)
	line64 += "\n"
	_, err := enc.wr.Write([]byte(line64))
	return err
}

// Meta lines are in JSON for readability.
// TODO: CRC not used (yet), concatenate the CRC bytes with the JSON bytes.
func (enc *WALWriter) WriteMeta(meta MetaMessage) error {
	metaJSON := amino.MustMarshalJSON(meta)
	metaLine := "#" + string(metaJSON) + "\n"
	_, err := enc.wr.Write([]byte(metaLine))
	return err
}

// -----------

// IsDataCorruptionError returns true if data has been corrupted inside WAL.
func IsDataCorruptionError(err error) bool {
	_, ok := err.(DataCorruptionError)
	return ok
}

// DataCorruptionError is an error that occurs if data on disk was corrupted.
type DataCorruptionError struct {
	cause error
}

func (e DataCorruptionError) Error() string {
	return fmt.Sprintf("DataCorruptionError[%v]", e.cause)
}

func (e DataCorruptionError) Cause() error {
	return e.cause
}

// A WALReader reads and decodes custom-encoded WAL messages from an input
// stream. See WALWriter for the format used.
//
// It will also compare the checksums and make sure data size is equal to the
// length from the header. If that is not the case, error will be returned.
//
// WALReader is itself an io.ReaderCloser, and it uses a bufio reader under the
// hood, which means it will usually end up reading more bytes than was
// actually returned via calls to ReadMessage().
type WALReader struct {
	rd      io.Reader // NOTE: use brd instead.
	brd     *bufio.Reader
	maxSize int64
}

// NewWALReader returns a new decoder that reads from rd.
func NewWALReader(rd io.Reader, maxSize int64) *WALReader {
	return &WALReader{rd, bufio.NewReader(rd), maxSize}
}

// Reads a line until the "\n" delimiter byte, and returns that line without
// the delimiter.
// A line must end with "\n", otherwise EOF.
func (dec *WALReader) readline() ([]byte, error) {
	bz, err := dec.brd.ReadBytes('\n')
	if 0 < len(bz) {
		bz = bz[:len(bz)-1]
	}
	return bz, err
}

// Implement io.ReadCloser for SearchForHeight() result reader.
func (dec *WALReader) Read(p []byte) (n int, err error) {
	return dec.brd.Read(p)
}

// Implement io.ReadCloser for SearchForHeight() result reader.
func (dec *WALReader) Close() (err error) {
	// There is no corresponding .Close on a bufio.
	// we will set brd to nil and let the program panic
	// when methods are called after a close.
	dec.brd = nil

	// Close rd if it is a Closer.
	if cl, ok := dec.rd.(io.Closer); ok {
		err = cl.Close()
		return
	}

	return
}

// Decode reads the next custom-encoded value from its reader and returns it.
// One TimedWALMessage or MetaError or error is returned, the rest are nil.
func (dec *WALReader) ReadMessage() (*TimedWALMessage, *MetaMessage, error) {
	line64, err := dec.readline()
	if err != nil {
		return nil, nil, err
	}

	if len(line64) == 0 {
		return nil, nil, DataCorruptionError{fmt.Errorf("found empty line")}
	}

	// special case for MetaMessage.
	if line64[0] == '#' {
		var meta MetaMessage
		err := amino.UnmarshalJSON(line64[1:], &meta)
		return nil, &meta, err
	}

	// is usual TimedWALMessage.
	// decode base64.
	line, err := base64stdnp.DecodeString(string(line64))
	if err != nil {
		return nil, nil, DataCorruptionError{fmt.Errorf("failed to decode base64: %w", err)}
	}

	// read crc out of bytes.
	crcSize := int64(4)
	if int64(len(line)) < crcSize {
		return nil, nil, DataCorruptionError{fmt.Errorf("failed to read checksum: %w", err)}
	}
	crc, twmBytes := binary.BigEndian.Uint32(line[:crcSize]), line[crcSize:]
	if dec.maxSize < int64(len(twmBytes)) {
		return nil, nil, DataCorruptionError{fmt.Errorf("length %d exceeded maximum possible value of %d bytes", int64(len(twmBytes)), dec.maxSize)}
	}

	// check checksum before decoding twmBytes
	if len(twmBytes) == 0 {
		return nil, nil, DataCorruptionError{fmt.Errorf("failed to read amino sized bytes: %w", err)}
	}
	actualCRC := crc32.Checksum(twmBytes, crc32c)
	if actualCRC != crc {
		return nil, nil, DataCorruptionError{fmt.Errorf("checksums do not match: read: %v, actual: %v", crc, actualCRC)}
	}

	// decode amino sized bytes.
	res := new(TimedWALMessage) //nolint: gosimple
	err = amino.UnmarshalSized(twmBytes, res)
	if err != nil {
		return nil, nil, DataCorruptionError{fmt.Errorf("failed to decode twmBytes: %w", err)}
	}

	return res, nil, err
}

type NopWAL struct{}

var _ WAL = NopWAL{}

func (NopWAL) SetLogger(l *slog.Logger)          {}
func (NopWAL) Write(m WALMessage) error          { return nil }
func (NopWAL) WriteSync(m WALMessage) error      { return nil }
func (NopWAL) WriteMetaSync(m MetaMessage) error { return nil }
func (NopWAL) FlushAndSync() error               { return nil }
func (NopWAL) SearchForHeight(height int64, options *WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
	return nil, false, nil
}
func (NopWAL) Start() error { return nil }
func (NopWAL) Stop() error  { return nil }
func (NopWAL) Wait()        {}
