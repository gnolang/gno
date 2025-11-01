package coverage

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Analyzer analyzes AST nodes to identify executable lines.
type Analyzer struct {
	tracker *Tracker
}

// NewAnalyzer creates a new analyzer with the given tracker.
func NewAnalyzer(tracker *Tracker) *Analyzer {
	return &Analyzer{tracker: tracker}
}

// AnalyzePackage analyzes a package to register all executable lines.
func (a *Analyzer) AnalyzePackage(pn *gnolang.PackageNode) {
	if pn == nil {
		return
	}

	for _, file := range pn.Files {
		a.analyzeFile(pn.PkgPath, file)
	}
}

// analyzeFile analyzes a single file to register executable lines.
func (a *Analyzer) analyzeFile(pkgPath string, fn *gnolang.FileNode) {
	if fn == nil {
		return
	}

	// Get file name from FileNode's location
	fileName := ""
	if loc := fn.GetLocation(); !loc.IsZero() {
		fileName = loc.File
	}

	// Analyze all declarations in the file
	for _, decl := range fn.Decls {
		a.analyzeDecl(pkgPath, fileName, decl)
	}
}

// analyzeDecl analyzes a declaration node.
func (a *Analyzer) analyzeDecl(pkgPath, fileName string, decl gnolang.Decl) {
	switch d := decl.(type) {
	case *gnolang.FuncDecl:
		// Analyze function body
		if d.Body != nil && len(d.Body) > 0 {
			for _, stmt := range d.Body {
				a.analyzeStmt(pkgPath, fileName, stmt)
			}
		}
	case *gnolang.ValueDecl:
		// Value declarations are typically not executable unless they have init expressions
		// We'll mark the line if there's an initialization
		if len(d.Values) > 0 && d.GetLine() > 0 {
			a.tracker.RegisterExecutableLine(pkgPath, fileName, d.GetLine())
		}
	}
}

// analyzeStmt analyzes a statement node to find executable lines.
func (a *Analyzer) analyzeStmt(pkgPath, fileName string, stmt gnolang.Stmt) {
	if stmt == nil {
		return
	}

	// Register this statement's line as executable
	if line := stmt.GetLine(); line > 0 {
		a.tracker.RegisterExecutableLine(pkgPath, fileName, line)
	}

	// Recursively analyze nested statements
	switch s := stmt.(type) {
	case *gnolang.BlockStmt:
		for _, innerStmt := range s.Body {
			a.analyzeStmt(pkgPath, fileName, innerStmt)
		}

	case *gnolang.IfStmt:
		if s.Init != nil {
			a.analyzeStmt(pkgPath, fileName, s.Init)
		}
		// Then is always present as IfCaseStmt
		for _, stmt := range s.Then.Body {
			a.analyzeStmt(pkgPath, fileName, stmt)
		}
		// Else might have a body
		if len(s.Else.Body) > 0 {
			for _, stmt := range s.Else.Body {
				a.analyzeStmt(pkgPath, fileName, stmt)
			}
		}

	case *gnolang.ForStmt:
		if s.Init != nil {
			a.analyzeStmt(pkgPath, fileName, s.Init)
		}
		if s.Post != nil {
			a.analyzeStmt(pkgPath, fileName, s.Post)
		}
		for _, bodyStmt := range s.Body {
			a.analyzeStmt(pkgPath, fileName, bodyStmt)
		}

	case *gnolang.RangeStmt:
		for _, bodyStmt := range s.Body {
			a.analyzeStmt(pkgPath, fileName, bodyStmt)
		}

	case *gnolang.SwitchStmt:
		if s.Init != nil {
			a.analyzeStmt(pkgPath, fileName, s.Init)
		}
		for _, clause := range s.Clauses {
			for _, bodyStmt := range clause.Body {
				a.analyzeStmt(pkgPath, fileName, bodyStmt)
			}
		}

	case *gnolang.SelectStmt:
		for _, scase := range s.Cases {
			for _, bodyStmt := range scase.Body {
				a.analyzeStmt(pkgPath, fileName, bodyStmt)
			}
		}

	case *gnolang.DeferStmt:
		// The defer statement itself is executable
		// The call will be tracked when executed

	case *gnolang.GoStmt:
		// The go statement itself is executable
		// The call will be tracked when executed

	// Expression statements are executable
	case *gnolang.ExprStmt:
		// Already registered above

	// Assignment statements are executable
	case *gnolang.AssignStmt:
		// Already registered above

	// Other simple statements are also executable
	case *gnolang.IncDecStmt:
		// Already registered above

	case *gnolang.ReturnStmt:
		// Already registered above

	case *gnolang.BranchStmt:
		// Already registered above

	case *gnolang.DeclStmt:
		// Declaration statements may contain initialization
		if len(s.Body) > 0 {
			// The declaration itself is tracked
		}
	}
}

// AnalyzeMemPackage analyzes a MemPackage to register all executable lines.
func (a *Analyzer) AnalyzeMemPackage(memPkg *std.MemPackage) {
	if memPkg == nil {
		return
	}

	// Note: MemPackage contains raw file content, not parsed AST
	// To properly analyze it, we would need to parse each file
	// This would require access to the Gno parser
	// For now, this is a placeholder for future implementation
}
