package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// PackageDiffChecker is a struct for comparing and identifying differences
// between files in two directories.
type PackageDiffChecker struct {
	SrcFiles []string // List of source files.
	SrcPath  string   // Source directory path.
	DstFiles []string // List of destination files.
	DstPath  string   // Destination directory path.
}

// Differences represents the differences between source and destination packages.
type Differences struct {
	SameNumberOfFiles bool             // Indicates whether the source and destination have the same number of files.
	FilesDifferences  []FileDifference // Differences in individual files.
}

// FileDifference represents the differences between source and destination files.
type FileDifference struct {
	Status          string            // Diff status of the processed files.
	SourceName      string            // Name of the source file.
	DestinationName string            // Name of the destination file.
	SrcLineDiff     []LineDifferrence // Differences in source file lines.
	DstLineDiff     []LineDifferrence // Differences in destination file lines.
}

// NewPackageDiffChecker creates a new PackageDiffChecker instance with the specified
// source and destination paths. It initializes the SrcFiles and DstFiles fields by
// listing files in the corresponding directories.
func NewPackageDiffChecker(srcPath, dstPath string) (*PackageDiffChecker, error) {
	srcFiles, err := listDirFiles(srcPath)
	if err != nil {
		return nil, err
	}

	dstFiles, err := listDirFiles(dstPath)
	if err != nil {
		return nil, err
	}

	return &PackageDiffChecker{
		SrcFiles: srcFiles,
		SrcPath:  srcPath,
		DstFiles: dstFiles,
		DstPath:  dstPath,
	}, nil
}

// Differences calculates and returns the differences between source and destination
// packages. It compares files line by line using the Myers algorithm.
func (p *PackageDiffChecker) Differences() (*Differences, error) {
	d := &Differences{
		SameNumberOfFiles: p.hasSameNumberOfFiles(),
		FilesDifferences:  make([]FileDifference, 0),
	}

	allFiles := p.listAllPossibleFiles()

	for _, trimmedFileName := range allFiles {
		srcFilePath := p.SrcPath + "/" + trimmedFileName + ".gno"
		dstFilePath := p.DstPath + "/" + trimmedFileName + ".go"

		fileDiff, err := NewFileDiff(srcFilePath, dstFilePath, "myers")
		if err != nil {
			return nil, err
		}

		srcDiff, dstDiff := fileDiff.Differences()

		d.FilesDifferences = append(d.FilesDifferences, FileDifference{
			Status:          p.getStatus(srcDiff, dstDiff).String(),
			SourceName:      trimmedFileName + ".gno",
			DestinationName: trimmedFileName + ".go",
			SrcLineDiff:     srcDiff,
			DstLineDiff:     dstDiff,
		})
	}

	return d, nil
}

// listAllPossibleFiles returns a list of unique file names without extensions
// from both source and destination directories.
func (p *PackageDiffChecker) listAllPossibleFiles() []string {
	files := p.SrcFiles
	files = append(files, p.DstFiles...)

	for i := 0; i < len(files); i++ {
		files[i] = strings.TrimSuffix(files[i], ".go")
		files[i] = strings.TrimSuffix(files[i], ".gno")
	}

	unique := make(map[string]bool, len(files))
	uniqueFiles := make([]string, len(unique))
	for _, file := range files {
		if len(file) != 0 {
			if !unique[file] {
				uniqueFiles = append(uniqueFiles, file)
				unique[file] = true
			}
		}
	}

	return uniqueFiles
}

func (p *PackageDiffChecker) getStatus(srcDiff, dstDiff []LineDifferrence) diffStatus {
	slicesAreEquals := slices.Equal(srcDiff, dstDiff)
	if slicesAreEquals {
		return NO_DIFF
	}

	if len(srcDiff) == 0 {
		return MISSING_IN_SRC
	}

	if len(dstDiff) == 0 {
		return MISSING_IN_DST
	}

	if !slicesAreEquals {
		return HAS_DIFF
	}

	return 0
}

// hasSameNumberOfFiles checks if the source and destination have the same number of files.
func (p *PackageDiffChecker) hasSameNumberOfFiles() bool {
	return len(p.SrcFiles) == len(p.DstFiles)
}

// listDirFiles returns a list of file names in the specified directory.
func listDirFiles(dirPath string) ([]string, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return []string{}, nil
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "can't close "+dirPath)
		}
	}()

	filesInfo, err := f.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("can't list file in directory :%w", err)
	}

	fileNames := make([]string, 0)
	for _, info := range filesInfo {
		fileNames = append(fileNames, info.Name())
	}

	return fileNames, nil
}
