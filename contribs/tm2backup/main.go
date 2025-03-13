package main

import (
	"archive/tar"
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"connectrpc.com/connect"
	backup "github.com/gnolang/gno/tm2/pkg/bft/node/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/node/backuppb/backupconnect"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gofrs/flock"
	"github.com/klauspost/compress/zstd"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cfg := &backupCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "[flags]",
			ShortHelp:  "efficiently backup tm2 blocks",
			LongHelp:   "Efficiently backup tendermint2 blocks",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execBackup(ctx, cfg, io)
		},
	)

	return cmd
}

type backupCfg struct {
	remote      string
	outDir      string
	startHeight uint64
	endHeight   uint64
}

func (c *backupCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"http://localhost:4242",
		"Backup service remote.",
	)
	fs.StringVar(
		&c.outDir,
		"o",
		"blocks-backup",
		"Output directory.",
	)
	fs.Uint64Var(
		&c.startHeight,
		"start",
		0,
		fmt.Sprintf("Start height. Will be aligned at a multiple of %d blocks. This can't be used when resuming from an existing output directory.", chunkSize),
	)
	fs.Uint64Var(
		&c.endHeight,
		"end",
		0,
		"End height, inclusive. Use 0 for latest height.",
	)
}

// XXX: versioning of output directory

const nextHeightFilename = "next-height.txt"
const chunkSize = 100

func execBackup(ctx context.Context, c *backupCfg, io commands.IO) error {
	if err := validateInput(c); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	unlock, err := lockOutputDir(c, io)
	if err != nil {
		return fmt.Errorf("failed to lock output directory: %w", err)
	}
	defer unlock()

	nextHeightFP := filepath.Join(c.outDir, nextHeightFilename)
	nextHeight, err := readNextHeight(nextHeightFP)
	if err != nil {
		return fmt.Errorf("failed to read next height: %w", err)
	}

	height, err := getStartHeight(int64(c.startHeight), nextHeight)
	if err != nil {
		return fmt.Errorf("failed to decide start height: %w", err)
	}

	if c.endHeight != 0 && int64(c.endHeight) < nextHeight {
		return fmt.Errorf("invalid input: requested end height is smaller than the next height in output directory (%d), use a different output directory or a valid end height", nextHeight)
	}

	prefix := "starting"
	if nextHeight != -1 {
		prefix = "resuming"
	}
	io.Println(prefix, "at height", height)

	client := backupconnect.NewBackupServiceClient(
		http.DefaultClient,
		c.remote,
		connect.WithGRPC(),
	)
	res, err := client.StreamBlocks(
		ctx,
		connect.NewRequest(&backup.StreamBlocksRequest{
			StartHeight: height,
			EndHeight:   int64(c.endHeight),
		}),
	)
	if err != nil {
		return err
	}

	chunkStart := height

	nextChunkFP := filepath.Join(c.outDir, "next-chunk.tar.zst")
	outFile, err := os.OpenFile(nextChunkFP, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o664)
	if err != nil {
		return err
	}
	zstw, err := zstd.NewWriter(outFile)
	if err != nil {
		return err
	}
	w := tar.NewWriter(zstw)

	finalizeChunk := func(last bool) error {
		if err := w.Close(); err != nil {
			return err
		}
		if err := zstw.Close(); err != nil {
			return err
		}
		if err := outFile.Close(); err != nil {
			return err
		}

		// using padding for filename to match chunk order and lexicographical order
		chunkFP := filepath.Join(c.outDir, fmt.Sprintf("%019d.tm2blocks.tar.zst", chunkStart))
		if err := os.Rename(nextChunkFP, chunkFP); err != nil {
			return err
		}

		if err := os.WriteFile(nextHeightFP, []byte(strconv.FormatInt(height, 10)), 0o664); err != nil {
			return err
		}

		io.Println("wrote blocks", chunkStart, "to", height-1, "at", chunkFP)

		if last {
			return nil
		}

		outFile, err = os.OpenFile(nextChunkFP, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o664)
		if err != nil {
			return err
		}
		zstw, err = zstd.NewWriter(outFile)
		if err != nil {
			return err
		}
		w = tar.NewWriter(zstw)

		chunkStart = height

		return nil
	}

	for {
		ok := res.Receive()
		if !ok {
			if res.Err() != nil {
				return err
			}

			return finalizeChunk(true)
		}
		msg := res.Msg()

		header := &tar.Header{
			Name: fmt.Sprintf("%d", msg.Height),
			Size: int64(len(msg.Data)),
			Mode: 0o664,
		}
		if err := w.WriteHeader(header); err != nil {
			return err
		}

		_, err := w.Write(msg.Data)
		if err != nil {
			return err
		}

		height += 1
		if height%chunkSize != 1 {
			continue
		}

		if err := finalizeChunk(false); err != nil {
			return err
		}

		// don't finalize twice if the requested endHeight is a multiple of chunkSize
		if height-1 == int64(c.endHeight) {
			return nil
		}
	}
}

func validateInput(c *backupCfg) error {
	if c.startHeight > math.MaxInt64 {
		return fmt.Errorf("start must be <= %d", math.MaxInt64)
	}

	if c.endHeight > math.MaxInt64 {
		return fmt.Errorf("end must be <= %d", math.MaxInt64)
	}

	return nil
}

func lockOutputDir(c *backupCfg, io commands.IO) (func(), error) {
	if err := os.MkdirAll(c.outDir, 0o775); err != nil {
		return nil, fmt.Errorf("failed to ensure output directory exists: %w", err)
	}

	fileLock := flock.New(filepath.Join(c.outDir, "blocks.lock"))
	locked, err := fileLock.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("failed to acquire lock on output directory")
	}
	return func() {
		if err := fileLock.Unlock(); err != nil {
			io.ErrPrintln(err)
		}
	}, nil
}

func getStartHeight(requestedStartHeight int64, outputDirNextHeight int64) (int64, error) {
	height := int64(1)

	if requestedStartHeight != 0 && outputDirNextHeight != -1 {
		return 0, errors.New("can't request a start height when resuming, use a different output directory or no start height")
	}

	if requestedStartHeight != 0 {
		height = requestedStartHeight
	} else {
		height = outputDirNextHeight
	}

	// align: 4 -> 1, 100 -> 1, 101 -> 101, 150 -> 101
	// we simply overwrite the latest chunk if it is partial because it's not expensive
	height -= (height - 1) % chunkSize

	if height < 1 || height%100 != 1 {
		return 0, fmt.Errorf("unexpected start height %d", height)
	}

	return height, nil
}

func readNextHeight(nextHeightFP string) (int64, error) {
	nextHeightBz, err := os.ReadFile(nextHeightFP)
	switch {
	case os.IsNotExist(err):
		return -1, nil
	case err != nil:
		return 0, err
	default:
		nextHeight, err := strconv.ParseInt(string(nextHeightBz), 10, 64)
		if err != nil {
			return 0, err
		}
		if nextHeight < 1 {
			return 0, fmt.Errorf("unexpected next height %d", nextHeight)
		}
		return nextHeight, nil
	}
}
