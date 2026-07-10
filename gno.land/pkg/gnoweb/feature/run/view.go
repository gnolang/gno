package run

import "path"

// RunData is the render payload for templates/page.html.
type RunData struct {
	PkgPath string // full path, e.g. "gno.land/r/demo/boards"
	Domain  string // e.g. "gno.land"
	Remote  string // e.g. "https://rpc.gno.land:443"
	ChainId string // e.g. "portal-loop"
}

// PkgAlias returns the last segment of the import path.
func (d RunData) PkgAlias() string {
	return path.Base(d.PkgPath)
}
