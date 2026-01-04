package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// DiffChecker is a struct for comparing differences between two files.
type DiffChecker struct {
	Src        string   // Name of the source file.
	Dst        string   // Name of the destination file.
	srcContent string   // Content of the source file.
	dstContent string   // Content of the destination file.
	srcLines   []string // Lines of the source file.
	dstLines   []string // Lines of the destination file.
}

// LineDifferrence represents a difference in a line during file comparison.
type LineDifferrence struct {
	SrcLine       string    // The line on Src.
	DestLine      string    // The line on Src.
	SrcOperation  operation // The operation performed on the line (e.g., "add", "delete", "equal").
	DestOperation operation
	SrcNumber     int
	DestNumber    int
}
type Diff struct {
	Diffs      []LineDifferrence
	MissingSrc bool
	MissingDst bool
}

// NewDiffChecker creates a new DiffChecker instance for comparing differences between
// the specified source and destination files. It initializes the source and
// destination file lines .
func NewDiffChecker(srcPath, dstPath string) (*DiffChecker, error) {
	src, err := getFileContent(srcPath)
	if err != nil {
		return nil, fmt.Errorf("can't read src file: %w", err)
	}

	dst, err := getFileContent(dstPath)
	if err != nil {
		return nil, fmt.Errorf("can't read dst file: %w", err)
	}

	return &DiffChecker{
		srcContent: src,
		dstContent: dst,
		srcLines:   strings.Split(src, "\n"),
		dstLines:   strings.Split(dst, "\n"),
		Src:        srcPath,
		Dst:        dstPath,
	}, nil
}

// Differences returns the differences in lines between the source and
// destination files using the configured diff algorithm.
func (f *DiffChecker) Differences() *Diff {
	var (
		srcIndex, dstIndex       int
		insertCount, deleteCount int
		diff                     []LineDifferrence
	)

	if len(f.dstContent) == 0 {
		return f.destEmpty()
	}

	if len(f.srcContent) == 0 {
		return f.srcEmpty()
	}

	/* printUntil prints all the lines than do not appear on the computed edits from gotextdiff
	so we need to add them manually looping always from the current value of
	srcIndex until the line before the start of the hunk computed diff, hunk.FromLine-1

	We need to print all the lines before each hunk and then ensure the end of the file is printed too
	*/
	printUntil := func(until int) {
		for i := srcIndex; i < until; i++ {
			diff = append(diff, LineDifferrence{
				SrcLine:       f.srcLines[srcIndex],
				DestLine:      f.srcLines[srcIndex],
				DestOperation: equal,
				SrcOperation:  equal,
				SrcNumber:     srcIndex + 1,
				DestNumber:    dstIndex + 1,
			})

			srcIndex++
			dstIndex++
		}
	}

	edits := myers.ComputeEdits(span.URIFromPath(f.Src), f.srcContent, f.dstContent)
	unified := gotextdiff.ToUnified(f.Src, f.Dst, f.srcContent, edits)
	for _, hunk := range unified.Hunks {
		printUntil(hunk.FromLine - 1)

		currentLine := LineDifferrence{}
		for _, line := range hunk.Lines {
			switch line.Kind {
			case gotextdiff.Insert:
				if currentLine.DestLine != "" {
					diff = append(diff, currentLine)
					currentLine = LineDifferrence{}
				}

				insertCount++
				dstIndex++

				currentLine.DestLine = line.Content
				currentLine.DestOperation = insert
				currentLine.DestNumber = dstIndex

			case gotextdiff.Equal:
				if currentLine.DestLine != "" || currentLine.SrcLine != "" {
					diff = append(diff, currentLine)
					currentLine = LineDifferrence{}
				}

				srcIndex++
				dstIndex++

				currentLine = LineDifferrence{
					SrcLine:       line.Content,
					DestLine:      line.Content,
					DestOperation: equal,
					SrcOperation:  equal,
					SrcNumber:     srcIndex,
					DestNumber:    dstIndex,
				}

			case gotextdiff.Delete:
				if currentLine.SrcLine != "" {
					diff = append(diff, currentLine)
					currentLine = LineDifferrence{}
				}
				srcIndex++
				deleteCount++
				currentLine.SrcLine = line.Content
				currentLine.SrcOperation = delete
				currentLine.SrcNumber = srcIndex
			}
		}
		diff = append(diff, currentLine)
	}

	printUntil(len(f.srcLines))

	return &Diff{
		Diffs: diff,
	}
}

func (f *DiffChecker) destEmpty() *Diff {
	diffs := []LineDifferrence{}
	for index, line := range f.srcLines {
		diffs = append(diffs, LineDifferrence{SrcLine: line, SrcOperation: delete, SrcNumber: index + 1})
	}

	return &Diff{
		Diffs:      diffs,
		MissingDst: true,
	}
}

func (f *DiffChecker) srcEmpty() *Diff {
	diffs := []LineDifferrence{}
	for index, line := range f.dstLines {
		diffs = append(diffs, LineDifferrence{DestLine: line, DestOperation: insert, DestNumber: index + 1})
	}

	return &Diff{
		Diffs:      diffs,
		MissingSrc: true,
	}
}

// getFileContent reads and returns the lines of a file given its path.
func getFileContent(p string) (string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.ReplaceAll(string(data), "\t", "    "), nil
}
