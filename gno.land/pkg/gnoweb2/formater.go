package gnoweb

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
)

type Formatter interface {
	Format(w io.Writer, iterator chroma.Iterator) error
}

type formaterWithStyle struct {
	*html.Formatter
	style *chroma.Style
}

func newFormaterWithStyle(formater *html.Formatter, style *chroma.Style) Formatter {
	return &formaterWithStyle{Formatter: formater, style: style}
}

func (f *formaterWithStyle) Format(w io.Writer, iterator chroma.Iterator) error {
	return f.Formatter.Format(w, f.style, iterator)
}
