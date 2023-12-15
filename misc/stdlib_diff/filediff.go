package main

import (
	"fmt"
	"os"
	"strings"
)

// FileDiff is a struct for comparing differences between two files.
type FileDiff struct {
	Src       []string // Lines of the source file.
	Dst       []string // Lines of the destination file.
	Algorithm          // Algorithm used for comparison.
}

// LineDifferrence represents a difference in a line during file comparison.
type LineDifferrence struct {
	Line      string    // The line content.
	Operation operation // The operation performed on the line (e.g., "add", "delete", "equal").
}

// NewFileDiff creates a new FileDiff instance for comparing differences between
// the specified source and destination files. It initializes the source and
// destination file lines .
func NewFileDiff(srcPath, dstPath string) (*FileDiff, error) {
	src, err := getFileLines(srcPath)
	if err != nil {
		return nil, fmt.Errorf("can't read src file: %w", err)
	}

	dst, err := getFileLines(dstPath)
	if err != nil {
		return nil, fmt.Errorf("can't read dst file: %w", err)
	}

	return &FileDiff{
		Src:       src,
		Dst:       dst,
		Algorithm: NewMyers(src, dst),
	}, nil
}

// Differences returns the differences in lines between the source and
// destination files using the configured diff algorithm.
func (f *FileDiff) Differences() (src, dst []LineDifferrence) {
	return f.Diff()
}

// getFileLines reads and returns the lines of a file given its path.
func getFileLines(p string) ([]string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	return lines, nil
}
