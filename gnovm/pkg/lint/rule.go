package lint

import "github.com/gnolang/gno/gnovm/pkg/gnolang"

type RuleCategory string

const (
	CategoryAVL     RuleCategory = "AVL"
	CategoryGeneral RuleCategory = "General"
	CategoryRender  RuleCategory = "Render"
)

type RuleInfo struct {
	ID       string
	Category RuleCategory
	Name     string
	Severity Severity
}

type RuleContext struct {
	PkgPath string
	File    *gnolang.FileNode
	Source  string
	Parents []gnolang.Node
}

type Rule interface {
	Info() RuleInfo
	Check(ctx *RuleContext, node gnolang.Node) []Issue
}
