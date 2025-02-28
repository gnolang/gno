package markdown

import (
	"github.com/yuin/goldmark"
)

type gno struct{}

// GnoExtension is an extension
var GnoExtension = &gno{}

func (e *gno) Extend(m goldmark.Markdown) {
	Column.Extend(m)
}
