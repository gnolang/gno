package markdown

import (
	"github.com/yuin/goldmark"
)

// GnoExtension is a goldmark Extender
var _ goldmark.Extender = (*gno)(nil)

type gno struct{}

// GnoExtension expose the gno extension, can be use with gno
var GnoExtension = &gno{}

func (e *gno) Extend(m goldmark.Markdown) {
	Column.Extend(m)
}
