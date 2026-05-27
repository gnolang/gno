package backup

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/klauspost/compress/zstd"
)

type Reader = func(yield func(block *types.Block) error) error

// WithReader creates a backup reader and pass it to the provided cb.
// yield should not be called concurently
func WithReader(dir string, startHeight int64, endHeight int64, cb func(reader Reader) error) (resErr error) {
	dir = filepath.Clean(dir)

	if endHeight != 0 && endHeight < startHeight {
		return fmt.Errorf("requested end height (%d) is smaller than requested start height (%d)", endHeight, startHeight)
	}

	unlock, err := lockDir(dir)
	if err != nil {
		return fmt.Errorf("lock output directory %q: %w", dir, err)
	}
	defer func() {
		resErr = errors.Join(unlock(), resErr)
	}()

	state, err := readState(dir)
	if err != nil {
		return err
	}

	if state.StartHeight < 1 || state.EndHeight < 1 {
		return errors.New("invalid backup state")
	}

	if startHeight < state.StartHeight {
		return fmt.Errorf("requested start height (#%d) is smaller than backup start height (#%d)", startHeight, state.StartHeight)
	}

	if endHeight == 0 {
		endHeight = state.EndHeight
	} else if endHeight > state.EndHeight {
		return fmt.Errorf("requested end height (#%d) is greater than backup end height (#%d)", endHeight, state.EndHeight)
	}

	chunkHeight := startHeight - (startHeight-1)%ChunkSize

	reader := &backupReader{
		dir:         dir,
		startHeight: startHeight,
		endHeight:   endHeight,
		chunkHeight: chunkHeight,
		height:      startHeight,
	}

	return cb(reader.read)
}

type backupReader struct {
	dir         string
	startHeight int64
	endHeight   int64
	chunkHeight int64
	height      int64
}

func (c *backupReader) read(yield func(block *types.Block) error) error {
	for {
		err := c.readChunk(yield)
		switch {
		case errors.Is(err, errReadDone):
			return nil
		case err != nil:
			return err
		}
	}
}

func (c *backupReader) readChunk(yield func(block *types.Block) error) (retErr error) {
	chunkFP := getChunkFP(c.dir, c.chunkHeight)

	chunk, err := os.Open(chunkFP)
	if err != nil {
		return err
	}
	defer func() {
		retErr = errors.Join(chunk.Close(), retErr)
	}()

	zstr, err := zstd.NewReader(chunk)
	if err != nil {
		return err
	}
	defer zstr.Close()

	r := tar.NewReader(zstr)
	for {
		h, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		blockHeight, err := strconv.ParseInt(h.Name, 10, 64)
		if err != nil {
			return err
		}

		if blockHeight < c.startHeight {
			continue
		}

		if blockHeight != c.height {
			return fmt.Errorf("unexpected block height: wanted %d, got %d", c.height, blockHeight)
		}

		blockBz := make([]byte, h.Size)
		if _, err := r.Read(blockBz); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		block := &types.Block{}
		if err := amino.Unmarshal(blockBz, block); err != nil {
			return err
		}

		if err := yield(block); err != nil {
			return err
		}

		if c.height == c.endHeight {
			return errReadDone
		}

		c.height += 1
	}

	c.chunkHeight += ChunkSize

	return nil
}

var errReadDone = errors.New("read done")
