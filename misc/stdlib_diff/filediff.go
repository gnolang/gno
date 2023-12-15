package main

import (
	"bufio"
	"os"
)

// FileDiff is a struct for comparing differences between two files.
type FileDiff struct {
	Src       []string // Lines of the source file.
	Dst       []string // Lines of the destination file.
	Algorithm          // Algorithm used for comparison.
}

// LineDifferrence represents a difference in a line during file comparison.
type LineDifferrence struct {
	Line      string // The line content.
	Operation string // The operation performed on the line (e.g., "add", "delete", "equal").
}

// NewFileDiff creates a new FileDiff instance for comparing differences between
// the specified source and destination files. It initializes the source and
// destination file lines .
func NewFileDiff(srcPath, dstPath string) (*FileDiff, error) {
	src := getFileLines(srcPath)
	dst := getFileLines(dstPath)

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
func getFileLines(p string) []string {
	lines := make([]string, 0)

	f, err := os.Open(p)
	if err != nil {
		return lines
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return lines
	}

	return lines
}
