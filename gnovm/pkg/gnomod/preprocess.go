package gnomod

import (
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func removeDups(syntax *modfile.FileSyntax, replace *[]*modfile.Replace) {
	if replace != nil {
		purged := removeReplaceDups(replace)
		cleanSyntaxTree(syntax, purged)
	}
}

// removeReplaceDups removes duplicate replacements.
// Later replacements take priority over earlier ones.
func removeReplaceDups(replace *[]*modfile.Replace) map[*modfile.Line]bool {
	purge := make(map[*modfile.Line]bool)

	haveReplace := make(map[module.Version]bool)
	for i := len(*replace) - 1; i >= 0; i-- {
		x := (*replace)[i]
		if haveReplace[x.Old] { // duplicate detected
			purge[x.Syntax] = true
			continue
		}
		haveReplace[x.Old] = true
	}
	var repl []*modfile.Replace
	for _, r := range *replace {
		if !purge[r.Syntax] {
			repl = append(repl, r)
		}
	}
	*replace = repl

	return purge
}

// cleanSyntaxTree removes purged statements from the syntax tree.
func cleanSyntaxTree(syntax *modfile.FileSyntax, purge map[*modfile.Line]bool) {
	stmts := make([]modfile.Expr, 0, len(syntax.Stmt))
	for _, stmt := range syntax.Stmt {
		switch stmt := stmt.(type) {
		case *modfile.Line:
			if purge[stmt] {
				continue
			}
		case *modfile.LineBlock:
			var lines []*modfile.Line
			for _, line := range stmt.Line {
				if !purge[line] {
					lines = append(lines, line)
				}
			}
			stmt.Line = lines
			if len(lines) == 0 {
				continue
			}
		}
		stmts = append(stmts, stmt)
	}
	syntax.Stmt = stmts
}
