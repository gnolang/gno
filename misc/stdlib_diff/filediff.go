package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// FileDiff is a struct for comparing differences between two files.
type FileDiff struct {
	Src        string   // Name of the source file.
	Dst        string   // Name of the destination file.
	srcContent string   // Content of the source file.
	dstContent string   // Content of the destination file.
	srcLines   []string // Lines of the source file.
	dstLines   []string // Lines of the destination file.

}

// LineDifferrence represents a difference in a line during file comparison.
type LineDifferrence struct {
	Line      string    // The line content.
	Operation operation // The operation performed on the line (e.g., "add", "delete", "equal").
	Number    int
}

// NewFileDiff creates a new FileDiff instance for comparing differences between
// the specified source and destination files. It initializes the source and
// destination file lines .
func NewFileDiff(srcPath, dstPath string) (*FileDiff, error) {
	src, err := getFileContent(srcPath)
	if err != nil {
		return nil, fmt.Errorf("can't read src file: %w", err)
	}

	dst, err := getFileContent(dstPath)
	if err != nil {
		return nil, fmt.Errorf("can't read dst file: %w", err)
	}

	return &FileDiff{
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
func (f *FileDiff) Differences() (src, dst []LineDifferrence) {
	var (
		srcIndex, dstIndex       int
		insertCount, deleteCount int
		dstDiff, srcDiff         []LineDifferrence
	)

	if len(f.dstContent) == 0 {
		return f.destEmpty()
	}

	if len(f.srcContent) == 0 {
		return f.srcEmpty()
	}

	/* printUntil prints all the lines thar are equal
	because they do not appear on the computed edits from gotextdiff
	so we need to add them manually looping always from the current value of
	srcIndex until the line before the start of the hunk computed diff, hunk.FromLine-1

	We need to print all the lines before each hunk and then ensure the end of the file is printed too
	*/
	printUntil := func(until int) {
		for i := srcIndex; i < until; i++ {
			dstDiff = append(dstDiff, LineDifferrence{Line: f.srcLines[srcIndex], Operation: equal, Number: dstIndex + 1})
			srcDiff = append(srcDiff, LineDifferrence{Line: f.srcLines[srcIndex], Operation: equal, Number: srcIndex + 1})
			srcIndex++
			dstIndex++
		}
	}

	edits := myers.ComputeEdits(span.URIFromPath(f.Src), f.srcContent, f.dstContent)
	unified := gotextdiff.ToUnified(f.Src, f.Dst, f.srcContent, edits)
	for _, hunk := range unified.Hunks {
		printUntil(hunk.FromLine - 1)

		for _, line := range hunk.Lines {
			switch line.Kind {
			case gotextdiff.Insert:
				insertCount++
				dstIndex++
				dstDiff = append(dstDiff, LineDifferrence{Line: line.Content, Operation: insert, Number: dstIndex})

			case gotextdiff.Equal:
				srcIndex++
				dstIndex++
				dstDiff = append(dstDiff, LineDifferrence{Line: line.Content, Operation: equal, Number: dstIndex})
				srcDiff = append(srcDiff, LineDifferrence{Line: line.Content, Operation: equal, Number: srcIndex})

			case gotextdiff.Delete:
				srcIndex++
				deleteCount++
				srcDiff = append(srcDiff, LineDifferrence{Line: line.Content, Operation: delete, Number: srcIndex})
			}
		}
	}

	printUntil(len(f.srcLines))
	return srcDiff, dstDiff
}

func (f *FileDiff) destEmpty() ([]LineDifferrence, []LineDifferrence) {
	srcDiff := []LineDifferrence{}
	for index, line := range f.srcLines {
		srcDiff = append(srcDiff, LineDifferrence{Line: line, Operation: delete, Number: index + 1})
	}

	return srcDiff, make([]LineDifferrence, 0)
}

func (f *FileDiff) srcEmpty() ([]LineDifferrence, []LineDifferrence) {
	destDiff := []LineDifferrence{}
	for index, line := range f.dstLines {
		destDiff = append(destDiff, LineDifferrence{Line: line, Operation: insert, Number: index + 1})
	}

	return make([]LineDifferrence, 0), destDiff
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
