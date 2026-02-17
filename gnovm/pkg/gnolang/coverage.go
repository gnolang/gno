package gnolang

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// CoverMode represents the coverage mode.
type CoverMode string

const (
	CoverModeSet    CoverMode = "set"
	CoverModeCount  CoverMode = "count"
	CoverModeAtomic CoverMode = "atomic"
)

// ParseCoverMode parses a cover mode string, returning an error for invalid modes.
func ParseCoverMode(s string) (CoverMode, error) {
	switch s {
	case "set":
		return CoverModeSet, nil
	case "count":
		return CoverModeCount, nil
	case "atomic":
		return CoverModeAtomic, nil
	default:
		return "", fmt.Errorf("invalid covermode %q: must be set, count, or atomic", s)
	}
}

// stmtInfo stores source location info for a tracked statement.
type stmtInfo struct {
	File      string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// StatementCoverage tracks covered statements for a package.
type StatementCoverage struct {
	Mode    CoverMode
	tracked map[uintptr]stmtInfo
	hits    map[uintptr]int
}

func NewStatementCoverage() *StatementCoverage {
	return NewStatementCoverageWithMode(CoverModeSet)
}

func NewStatementCoverageWithMode(mode CoverMode) *StatementCoverage {
	return &StatementCoverage{
		Mode:    mode,
		tracked: make(map[uintptr]stmtInfo, 256),
		hits:    make(map[uintptr]int, 256),
	}
}

func statementID(s Stmt) uintptr {
	return reflect.ValueOf(s).Pointer()
}

func shouldTrackStmt(s Stmt) bool {
	switch s.(type) {
	case *BlockStmt, *EmptyStmt, *IfCaseStmt, *SwitchClauseStmt, *bodyStmt:
		return false
	default:
		return true
	}
}

func (c *StatementCoverage) TrackStmt(s Stmt) {
	if c == nil || !shouldTrackStmt(s) {
		return
	}
	id := statementID(s)
	span := s.GetSpan()
	c.tracked[id] = stmtInfo{
		StartLine: span.Pos.Line,
		StartCol:  span.Pos.Column,
		EndLine:   span.End.Line,
		EndCol:    span.End.Column,
	}
}

// TrackStmtWithFile tracks a statement and associates it with a filename.
func (c *StatementCoverage) TrackStmtWithFile(s Stmt, file string) {
	if c == nil || !shouldTrackStmt(s) {
		return
	}
	id := statementID(s)
	span := s.GetSpan()
	c.tracked[id] = stmtInfo{
		File:      file,
		StartLine: span.Pos.Line,
		StartCol:  span.Pos.Column,
		EndLine:   span.End.Line,
		EndCol:    span.End.Column,
	}
}

func (c *StatementCoverage) TrackNode(n Node) {
	c.TrackNodeWithFile(n, "")
}

// TrackNodeWithFile walks the AST and tracks all statements with the given filename.
func (c *StatementCoverage) TrackNodeWithFile(n Node, file string) {
	if c == nil || n == nil {
		return
	}
	_ = Transcribe(n, func(_ []Node, _ TransField, _ int, cn Node, stage TransStage) (Node, TransCtrl) {
		if stage == TRANS_ENTER {
			if s, ok := cn.(Stmt); ok {
				c.TrackStmtWithFile(s, file)
			}
		}
		return cn, TRANS_CONTINUE
	})
}

func (c *StatementCoverage) MarkExecuted(s Stmt) {
	if c == nil || !shouldTrackStmt(s) {
		return
	}
	id := statementID(s)
	if _, ok := c.tracked[id]; !ok {
		return
	}
	switch c.Mode {
	case CoverModeCount, CoverModeAtomic:
		c.hits[id]++
	default: // set
		c.hits[id] = 1
	}
}

func (c *StatementCoverage) Percent() float64 {
	if c == nil || len(c.tracked) == 0 {
		return 0
	}
	return float64(len(c.hits)) * 100 / float64(len(c.tracked))
}

// HitCount returns the execution count for a statement, or 0 if not hit.
func (c *StatementCoverage) HitCount(s Stmt) int {
	if c == nil {
		return 0
	}
	return c.hits[statementID(s)]
}

// ProfileEntry represents a single line in a Go cover profile.
type ProfileEntry struct {
	FileName  string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}

// Profile generates a Go cover profile string compatible with `go tool cover`.
func (c *StatementCoverage) Profile(pkgPath string) string {
	if c == nil || len(c.tracked) == 0 {
		return ""
	}

	mode := string(c.Mode)
	if mode == "" {
		mode = "set"
	}

	var entries []ProfileEntry
	for id, info := range c.tracked {
		fname := info.File
		if fname == "" {
			fname = "unknown.gno"
		}
		// Use pkgPath/filename as the Go cover profile expects.
		fullPath := pkgPath + "/" + fname

		count := c.hits[id]
		entries = append(entries, ProfileEntry{
			FileName:  fullPath,
			StartLine: info.StartLine,
			StartCol:  info.StartCol,
			EndLine:   info.EndLine,
			EndCol:    info.EndCol,
			NumStmt:   1,
			Count:     count,
		})
	}

	// Sort for deterministic output.
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.FileName != b.FileName {
			return a.FileName < b.FileName
		}
		if a.StartLine != b.StartLine {
			return a.StartLine < b.StartLine
		}
		return a.StartCol < b.StartCol
	})

	var sb strings.Builder
	fmt.Fprintf(&sb, "mode: %s\n", mode)
	for _, e := range entries {
		fmt.Fprintf(&sb, "%s:%d.%d,%d.%d %d %d\n",
			e.FileName, e.StartLine, e.StartCol, e.EndLine, e.EndCol,
			e.NumStmt, e.Count)
	}
	return sb.String()
}
