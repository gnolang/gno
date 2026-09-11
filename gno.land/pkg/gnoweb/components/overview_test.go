package components

import (
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// weburlParseForTest wraps weburl.Parse for test readability.
func weburlParseForTest(path string) (*weburl.GnoURL, error) {
	return weburl.Parse(path)
}

func fileContentFn(m map[string][]byte) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		v, ok := m[name]
		return v, ok
	}
}

// noopRenderer renders doc strings by writing them unchanged — enough for unit tests.
type noopRenderer struct{}

func (noopRenderer) RenderDocumentation(w io.Writer, src []byte) error {
	_, err := w.Write(src)
	return err
}

func (noopRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	_, err := w.Write(src)
	return err
}
