package gnoweb

import (
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

type RealmMeta struct {
	Toc md.Toc
}

// Renderer is an interface for rendering content from source.
type ContentRenderer interface {
	// Render renders the content of a source file and write it on the given writer.
	// It returns a Table of Contents (Toc) and an error if any occurs.
	RenderRealm(u *weburl.GnoURL, src []byte) (md.Toc, error)

	RenderSource(filename string, src []byte) error
}
