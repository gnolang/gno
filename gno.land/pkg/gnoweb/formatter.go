package gnoweb

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
)

type Formatter interface {
	Format(w io.Writer, iterator chroma.Iterator) error
}

type formatterWithStyle struct {
	*html.Formatter
	style *chroma.Style
}

func newFormatterWithStyle(formater *html.Formatter, style *chroma.Style) Formatter {
	return &formatterWithStyle{Formatter: formater, style: style}
}

func (f *formatterWithStyle) Format(w io.Writer, iterator chroma.Iterator) error {
	return f.Formatter.Format(w, f.style, iterator)
}
