package backup

import (
	"archive/tar"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/klauspost/compress/zstd"
	"go.uber.org/zap"
)

const (
	archiveSuffix     = ".tar.zst"
	nextChunkFilename = "next-chunk" + archiveSuffix
)

type Writer = func(block *types.Block) error

// WithWriter creates a backup writer and pass it to the provided cb.
// It is process-safe but not thread-safe.
func WithWriter(dir string, startHeightReq int64, endHeight int64, logger *zap.Logger, cb func(startHeight int64, write Writer) error) (retErr error) {
	if startHeightReq < 0 {
		return errors.New("start height request must be >= 0")
	}
	if endHeight < 0 {
		return errors.New("end height must be >= 0")
	}

	dir = filepath.Clean(dir)

	unlock, err := lockDir(dir)
	if err != nil {
		return fmt.Errorf("lock output directory %q: %w", dir, err)
	}
	defer func() {
		retErr = errors.Join(unlock(), retErr)
	}()

	state, err := readState(dir)
	if err != nil {
		return err
	}

	if endHeight != 0 && int64(endHeight) == state.EndHeight {
		logger.Info("Nothing to do, backup is already at requested end height")
		return nil
	}

	if endHeight != 0 && int64(endHeight) <= state.EndHeight {
		return fmt.Errorf("invalid input: requested end height is smaller or equal to the existing backup height (#%d), use a different output directory or a valid end height", state.EndHeight)
	}

	height, err := getStartHeight(int64(startHeightReq), state.EndHeight)
	if err != nil {
		return fmt.Errorf("decide start height: %w", err)
	}

	state.StartHeight = height

	prefix := "Starting backup"
	if state.EndHeight != -1 {
		prefix = "Resuming backup"
	}
	logger.Info(prefix, zap.Int64("height", height), zap.String("dir", dir), zap.Int64("end", endHeight))

	writer := writerImpl{
		dir:        dir,
		chunkStart: height,
		nextHeight: height,
		logger:     logger,
		state:      state,
	}

	if err := writer.openChunk(); err != nil {
		return fmt.Errorf("open first chunk: %w", err)
	}
	defer func() {
		retErr = errors.Join(writer.finalizeChunk(), retErr)
	}()

	return cb(height, writer.write)
}

type writerImpl struct {
	dir        string
	nextHeight int64
	chunkStart int64
	outFile    *os.File
	logger     *zap.Logger
	zstw       *zstd.Encoder
	w          *tar.Writer
	poisoned   bool
	state      *backupState
}

func (b *writerImpl) write(block *types.Block) error {
	if b.poisoned {
		return errors.New("poisoned")
	}

	if block.Height != b.nextHeight {
		return fmt.Errorf("non-contiguous block received, expected height #%d, got #%d", b.nextHeight, block.Height)
	}

	blockBz, err := amino.Marshal(block)
	if err != nil {
		return fmt.Errorf("marshal block: %w", err)
	}

	header := tar.Header{
		Name: fmt.Sprintf("%d", b.nextHeight),
		Size: int64(len(blockBz)),
		Mode: 0o664,
	}
	if err := b.w.WriteHeader(&header); err != nil {
		b.poisoned = true
		return fmt.Errorf("write tar file header: %w", err)
	}

	if _, err := b.w.Write(blockBz); err != nil {
		b.poisoned = true
		return fmt.Errorf("write block body: %w", err)
	}

	b.nextHeight += 1
	if b.nextHeight%ChunkSize != 1 {
		return nil
	}

	if err := b.finalizeChunk(); err != nil {
		b.poisoned = true
		return fmt.Errorf("finalize chunk: %w", err)
	}

	b.chunkStart = b.nextHeight

	if err := b.openChunk(); err != nil {
		b.poisoned = true
		return fmt.Errorf("open chunk: %w", err)
	}

	return nil
}

func (b *writerImpl) openChunk() error {
	var err error

	nextChunkFP := filepath.Join(b.dir, nextChunkFilename)

	b.outFile, err = os.OpenFile(nextChunkFP, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o664)
	if err != nil {
		return err
	}

	b.zstw, err = zstd.NewWriter(b.outFile)
	if err != nil {
		return errors.Join(b.outFile.Close(), err)
	}

	b.w = tar.NewWriter(b.zstw)

	return nil
}

func (b *writerImpl) finalizeChunk() error {
	if err := b.closeWriters(); err != nil {
		return err
	}

	if b.nextHeight == b.chunkStart || b.poisoned {
		return nil
	}

	nextChunkFP := filepath.Join(b.dir, nextChunkFilename)

	// using padding for filename to match chunk order and lexicographical order
	chunkFP := getChunkFP(b.dir, b.chunkStart)
	if err := os.Rename(nextChunkFP, chunkFP); err != nil {
		return err
	}

	b.state.EndHeight = b.nextHeight - 1
	if err := b.state.save(); err != nil {
		return err
	}

	b.logger.Debug("Wrote chunk", zap.Int64("start", b.chunkStart), zap.Int64("end", b.nextHeight-1), zap.String("file", chunkFP))

	return nil
}

func (b *writerImpl) closeWriters() error {
	errs := make([]error, 0, 3)

	if b.w != nil {
		errs = append(errs, b.w.Close())
		b.w = nil
	}

	if b.zstw != nil {
		errs = append(errs, b.zstw.Close())
		b.zstw = nil
	}

	if b.outFile != nil {
		errs = append(errs, b.outFile.Close())
		b.outFile = nil
	}

	return errors.Join(errs...)
}

func getStartHeight(requestedStartHeight int64, backupHeight int64) (int64, error) {
	height := int64(1)

	if requestedStartHeight != 0 && backupHeight != -1 {
		return 0, errors.New("can't request a start height when resuming, use a different output directory or no start height")
	}

	if requestedStartHeight != 0 {
		height = requestedStartHeight
	} else {
		height = backupHeight + 1
	}

	// align: 4 -> 1, 100 -> 1, 101 -> 101, 150 -> 101 (with ChunkSize == 100)
	// we simply overwrite the latest chunk if it is partial because it's not expensive
	height -= (height - 1) % ChunkSize

	if height < 1 || height%ChunkSize != 1 {
		return 0, fmt.Errorf("unexpected start height %d", height)
	}

	return height, nil
}
