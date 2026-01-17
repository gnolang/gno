package rules

import (
	"unicode"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/lint"
)

type GLOBAL001 struct{}

func init() {
	lint.MustRegister(&GLOBAL001{})
}

func (GLOBAL001) Info() lint.RuleInfo {
	return lint.RuleInfo{
		ID:       "GLOBAL001",
		Category: lint.CategoryGeneral,
		Name:     "exported-global-variable",
		Severity: lint.SeverityWarning,
	}
}

func (GLOBAL001) Check(ctx *lint.RuleContext, node gnolang.Node) []lint.Issue {
	decl, ok := node.(*gnolang.ValueDecl)
	if !ok {
		return nil
	}

	if decl.Const {
		return nil
	}

	if !isFileLevelDecl(ctx.Parents) {
		return nil
	}

	issues := make([]lint.Issue, 0, len(decl.NameExprs))

	for _, nx := range decl.NameExprs {
		name := string(nx.Name)

		if name == "_" {
			continue
		}

		if !isExported(name) {
			continue
		}

		issues = append(issues, lint.NewIssue(
			"GLOBAL001",
			lint.SeverityWarning,
			"exported package-level variable: "+name,
			ctx.File.FileName,
			nx.GetPos(),
		))
	}

	return issues
}

func isFileLevelDecl(parents []gnolang.Node) bool {
	if len(parents) == 0 {
		return false
	}
	_, ok := parents[len(parents)-1].(*gnolang.FileNode)
	return ok
}

func isExported(name string) bool {
	for _, r := range name {
		return unicode.IsUpper(r)
	}
	return false
}
