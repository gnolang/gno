package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

const (
	ChunkSize           = 100
	heightFilename      = "next-height.txt"
	backupStateFilename = "info.json"
	backupVersion       = "v1"
)

func lockDir(dir string) (func() error, error) {
	if err := os.MkdirAll(dir, 0o775); err != nil {
		return nil, fmt.Errorf("failed to ensure output directory exists: %w", err)
	}

	fileLock := flock.New(filepath.Join(dir, "blocks.lock"))
	locked, err := fileLock.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("failed to acquire lock on output directory")
	}
	return fileLock.Unlock, nil
}

func getChunkFP(dir string, chunkHeight int64) string {
	return filepath.Join(dir, fmt.Sprintf("%019d.tm2blocks"+archiveSuffix, chunkHeight))
}

type backupState struct {
	Version     string
	StartHeight int64
	EndHeight   int64

	filepath string
}

func (s *backupState) save() error {
	bz, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.filepath, bz, 0o664); err != nil {
		return err
	}

	return nil
}

func readState(dir string) (*backupState, error) {
	fp := filepath.Join(dir, backupStateFilename)

	stateBz, err := os.ReadFile(fp)
	switch {
	case os.IsNotExist(err):
		return &backupState{
			StartHeight: -1,
			EndHeight:   -1,
			Version:     backupVersion,
			filepath:    fp,
		}, nil
	case err != nil:
		return nil, err
	}

	state := backupState{}
	if err := json.Unmarshal(stateBz, &state); err != nil {
		return nil, err
	}

	if state.Version != backupVersion {
		return nil, fmt.Errorf("backup version mismatch, expected %q (binary), got %q (directory)", backupVersion, state.Version)
	}

	state.filepath = fp

	return &state, nil
}
