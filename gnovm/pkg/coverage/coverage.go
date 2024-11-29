package coverage

import (
	"io"
	"path/filepath"
)

// Collector defines the interface for collecting coverage data
type Collector interface {
	RecordHit(loc FileLocation)
	SetExecutableLines(filePath string, lines map[int]bool)
	AddFile(filePath string, totalLines int)
}

// Coverage implements the Collector interface and manages coverage data
type Coverage struct {
	enabled     bool
	rootDir     string
	currentPath string
	currentFile string
	files       fileCoverageMap
	pathCache   pathCache
}

// FileCoverage stores coverage information for a single file
type FileCoverage struct {
	totalLines      int
	hitLines        map[int]int
	executableLines map[int]bool
}

type (
	fileCoverageMap map[string]FileCoverage
	pathCache       map[string]string
)

func (m fileCoverageMap) get(path string) (FileCoverage, bool) {
	fc, ok := m[path]
	return fc, ok
}

func (m fileCoverageMap) set(path string, fc FileCoverage) {
	m[path] = fc
}

// NewFileCoverage creates a new FileCoverage instance
func NewFileCoverage() FileCoverage {
	return FileCoverage{
		totalLines:      0,
		hitLines:        make(map[int]int),
		executableLines: make(map[int]bool),
	}
}

// New creates a new Coverage instance
func New(rootDir string) *Coverage {
	return &Coverage{
		rootDir:   rootDir,
		files:     make(fileCoverageMap),
		pathCache: make(pathCache),
	}
}

// Configuration methods
func (c *Coverage) Enabled() bool              { return c.enabled }
func (c *Coverage) Enable()                    { c.enabled = true }
func (c *Coverage) Disable()                   { c.enabled = false }
func (c *Coverage) SetCurrentPath(path string) { c.currentPath = path }
func (c *Coverage) CurrentPath() string        { return c.currentPath }
func (c *Coverage) SetCurrentFile(file string) { c.currentFile = file }
func (c *Coverage) CurrentFile() string        { return c.currentFile }

// RecordHit implements Collector.RecordHit
func (c *Coverage) RecordHit(loc FileLocation) {
	if !c.enabled { return }

	path := filepath.Join(loc.PkgPath, loc.File)
	cov := c.getOrCreateFileCoverage(path)

	if cov.executableLines[loc.Line] {
		cov.hitLines[loc.Line]++
		c.files.set(path, cov)
	}
}

// SetExecutableLines implements Collector.SetExecutableLines
func (c *Coverage) SetExecutableLines(filePath string, executableLines map[int]bool) {
	cov, exists := c.files.get(filePath)
	if !exists {
		cov = NewFileCoverage()
	}

	cov.executableLines = executableLines
	c.files.set(filePath, cov)
}

// AddFile implements Collector.AddFile
func (c *Coverage) AddFile(filePath string, totalLines int) {
	if IsTestFile(filePath) || !isValidFile(c.currentPath, filePath) {
		return
	}

	cov, exists := c.files.get(filePath)
	if !exists {
		cov = NewFileCoverage()
	}

	cov.totalLines = totalLines
	c.files.set(filePath, cov)
}

// Report generates a coverage report using the given options
func (c *Coverage) Report(opts ReportOpts, w io.Writer) error {
	reporter := NewReporter(c, opts)
	if opts.pattern != "" {
		return reporter.WriteFileDetail(w, opts.pattern, opts.showHits)
	}
	return reporter.Write(w)
}

// FileLocation represents a specific location in source code
type FileLocation struct {
	PkgPath string
	File    string
	Line    int
	Column  int
}

// Helper methods
func (c *Coverage) getOrCreateFileCoverage(filePath string) FileCoverage {
	cov, exists := c.files.get(filePath)
	if !exists {
		cov = NewFileCoverage()
	}
	return cov
}

// GetStats returns coverage statistics for a file
func (c *Coverage) GetStats(filePath string) (hits, total int, ok bool) {
	cov, exists := c.files.get(filePath)
	if !exists {
		return 0, 0, false
	}
	return len(cov.hitLines), cov.totalLines, true
}

// GetFileHits returns the hit counts for a file (primarily for testing)
func (c *Coverage) GetFileHits(filePath string) map[int]int {
	if cov, exists := c.files.get(filePath); exists {
		return cov.hitLines
	}
	return nil
}

// GetExecutableLines returns the executable lines for a file (primarily for testing)
func (c *Coverage) GetExecutableLines(filePath string) map[int]bool {
	if cov, exists := c.files.get(filePath); exists {
		return cov.executableLines
	}
	return nil
}

// GetFiles returns a list of all tracked files
func (c *Coverage) GetFiles() []string {
	files := make([]string, 0, len(c.files))
	for file := range c.files {
		files = append(files, file)
	}
	return files
}
