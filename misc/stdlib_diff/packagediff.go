package main

import (
	"os"
	"path/filepath"
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
	Status            string            // Diff status of the processed files.
	SourceName        string            // Name of the source file.
	DestinationName   string            // Name of the destination file.
	LineDiffferrences []LineDifferrence // Differences in source file lines.
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

	srcFilesExt, dstFileExt := p.inferFileExtensions()
	allFiles := p.listAllPossibleFiles()

	for _, trimmedFileName := range allFiles {
		srcFileName := trimmedFileName + srcFilesExt
		srcFilePath := p.SrcPath + "/" + srcFileName
		dstFileName := trimmedFileName + dstFileExt
		dstFilePath := p.DstPath + "/" + dstFileName

		diffChecker, err := NewDiffChecker(srcFilePath, dstFilePath)
		if err != nil {
			return nil, err
		}

		diff := diffChecker.Differences()

		d.FilesDifferences = append(d.FilesDifferences, FileDifference{
			Status:            p.getStatus(diff).String(),
			SourceName:        srcFileName,
			DestinationName:   dstFileName,
			LineDiffferrences: diff.Diffs,
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

// inferFileExtensions by returning the src and dst files extensions.
func (p *PackageDiffChecker) inferFileExtensions() (string, string) {
	var goFiles, gnoFiles int
	for _, file := range p.SrcFiles {
		switch filepath.Ext(file) {
		case ".go":
			goFiles++
		case ".gno":
			gnoFiles++
		}
	}
	if goFiles > gnoFiles {
		return ".go", ".gno"
	}

	return ".gno", ".go"
}

// getStatus determines the diff status based on the differences in source and destination.
// It returns a diffStatus indicating whether there is no difference, missing in source, missing in destination, or differences exist.
func (p *PackageDiffChecker) getStatus(diff *Diff) diffStatus {
	if diff.MissingSrc {
		return missingInSrc
	}

	if diff.MissingDst {
		return missingInDst
	}

	for _, diff := range diff.Diffs {
		if diff.SrcOperation == delete || diff.DestOperation == insert {
			return hasDiff
		}
	}

	return noDiff
}

// hasSameNumberOfFiles checks if the source and destination have the same number of files.
func (p *PackageDiffChecker) hasSameNumberOfFiles() bool {
	return len(p.SrcFiles) == len(p.DstFiles)
}

// listDirFiles returns a list of file names in the specified directory.
func listDirFiles(dirPath string) ([]string, error) {
	fileNames := make([]string, 0)
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}
		// Only list .go and .gno files
		if !strings.Contains(dirEntry.Name(), ".go") && !strings.Contains(dirEntry.Name(), ".gno") {
			continue
		}
		fileNames = append(fileNames, dirEntry.Name())
	}

	return fileNames, nil
}
