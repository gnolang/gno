package run

import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"

// RunData is the render payload for templates/page.html.
type RunData struct {
	PkgPath string // full path, e.g. "gno.land/r/demo/boards"
	Remote  string // e.g. "https://rpc.gno.land:443"
	ChainId string // e.g. "portal-loop"
}

// PkgAlias returns the name an unaliased import of PkgPath binds: the last
// path element, skipping a version suffix ("gno.land/r/foo/v2" gives "foo").
func (d RunData) PkgAlias() string {
	return gno.LastPathElement(d.PkgPath)
}
