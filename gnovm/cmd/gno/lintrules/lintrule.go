package lintrules

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// The PoC aim to implement a way where gnolang pkg does not about linter
// We use TranscribeB to go through the AST built with gnolang.Nodes
// And we apply all the LintRules activated on all nodes.

type RuleContext struct {
	Store  gnolang.Store
	File   *gnolang.FileNode
	Source string
}

type LintRule interface {
	Run(ctx *RuleContext, node gnolang.Node) error
}
