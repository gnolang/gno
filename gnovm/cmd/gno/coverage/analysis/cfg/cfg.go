package cfg

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// CFGBuilder defines the interface for building a control flow graph (CFG).
type CFGBuilder interface {
	Build(stmts []ast.Stmt) *CFG
	Sort(stmts []ast.Stmt)
	PrintDot(f io.Writer, fset *token.FileSet, addl func(n ast.Stmt) string)
}

// CFG defines a control flow graph with statement-level granularity, in which
// there is a 1-1 correspondence between a block in the CFG and an ast.Stmt.
type CFG struct {
	// Sentinel nodes for single-entry CFG. Not in original AST.
	Entry *ast.BadStmt

	// Sentinel nodes for single-exit CFG. Not in original AST.
	Exit *ast.BadStmt

	// All defers found in CFG, disjoint from blocks. May be flowed to after Exit.
	Defers []*ast.DeferStmt
	blocks map[ast.Stmt]*block
}

type block struct {
	stmt  ast.Stmt
	preds []ast.Stmt
	succs []ast.Stmt
}

// FromStmts returns the control-flow graph for the given sequence of statements.
func FromStmts(s []ast.Stmt) *CFG {
	return NewBuilder().Build(s)
}

// FromFunc is a convenience function for creating a CFG from a given function declaration.
func FromFunc(f *ast.FuncDecl) *CFG {
	return FromStmts(f.Body.List)
}

// Preds returns a slice of all immediate predecessors for the given statement.
// May include Entry node.
func (c *CFG) Preds(s ast.Stmt) []ast.Stmt {
	return c.blocks[s].preds
}

// Succs returns a slice of all immediate successors to the given statement.
// May include Exit node.
func (c *CFG) Succs(s ast.Stmt) []ast.Stmt {
	return c.blocks[s].succs
}

// Blocks returns a slice of all blocks in a CFG, including the Entry and Exit nodes.
// The blocks are roughly in the order they appear in the source code.
func (c *CFG) Blocks() []ast.Stmt {
	blocks := make([]ast.Stmt, 0, len(c.blocks))
	for s := range c.blocks {
		blocks = append(blocks, s)
	}
	return blocks
}

// type for sorting statements by their starting positions in the source code
type stmtSlice []ast.Stmt

func (n stmtSlice) Len() int      { return len(n) }
func (n stmtSlice) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n stmtSlice) Less(i, j int) bool {
	return n[i].Pos() < n[j].Pos()
}

func (c *CFG) Sort(stmts []ast.Stmt) {
	sort.Sort(stmtSlice(stmts))
}

func (c *CFG) PrintDot(f io.Writer, fset *token.FileSet, addl func(n ast.Stmt) string) {
	fmt.Fprintf(f, `digraph mgraph {
mode="heir";
splines="ortho";

`)
	blocks := c.Blocks()
	c.Sort(blocks)
	for _, from := range blocks {
		succs := c.Succs(from)
		c.Sort(succs)
		for _, to := range succs {
			fmt.Fprintf(f, "\t\"%s\" -> \"%s\"\n",
				c.printVertex(from, fset, addl(from)),
				c.printVertex(to, fset, addl(to)))
		}
	}
	fmt.Fprintf(f, "}\n")
}

func (c *CFG) printVertex(stmt ast.Stmt, fset *token.FileSet, addl string) string {
	switch stmt {
	case c.Entry:
		return "ENTRY"
	case c.Exit:
		return "EXIT"
	case nil:
		return ""
	}
	addl = strings.Replace(addl, "\n", "\\n", -1)
	if addl != "" {
		addl = "\\n" + addl
	}
	return fmt.Sprintf("%s - line %d%s",
		astutil.NodeDescription(stmt),
		fset.Position(stmt.Pos()).Line,
		addl)
}
