package lintrules

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type RuleContext struct {
	Store  gnolang.Store
	File   *gnolang.FileNode
	Source string
}

type LintRule interface {
	Run(ctx *RuleContext, node gnolang.Node) error
}
