package lint

import "github.com/gnolang/gno/gnovm/pkg/gnolang"

type RuleCategory string

const (
	CategoryAVL     RuleCategory = "AVL"
	CategoryGeneral RuleCategory = "General"
)

type RuleInfo struct {
	ID       string
	Category RuleCategory
	Name     string
	Severity Severity
	// Description string // deferred
	// Link        string // deferred
}

type RuleContext struct {
	File    *gnolang.FileNode
	Source  string
	Parents []gnolang.Node // parent node stack (innermost last)
	// Store  gnolang.Store          // if rules need dynamic type resolution
	// Package *gnolang.PackageNode   // if rules need package-level info
	// Config  map[string]interface{} // for ConfigurableRule
}

type Rule interface {
	Info() RuleInfo
	Check(ctx *RuleContext, node gnolang.Node) []Issue
}

// ConfigurableRule interface deferred - add when rules need per-rule config
// type ConfigurableRule interface {
// 	Rule
// 	Init(cfg map[string]interface{}) error
// }
