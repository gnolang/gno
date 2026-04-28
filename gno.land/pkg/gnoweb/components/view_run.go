package components

import "path"

const RunViewType ViewType = "run-view"

// RunData holds the data for the maketx-run scratchpad view.
type RunData struct {
	PkgPath string // full path, e.g. "gno.land/r/demo/boards"
	Domain  string // e.g. "gno.land"
	Remote  string // e.g. "https://rpc.gno.land:443"
	ChainId string // e.g. "portal-loop"
}

// PkgAlias returns the last segment of the import path, used as the package alias
// in the generated template code (e.g. "boards" from "gno.land/r/demo/boards").
func (d RunData) PkgAlias() string {
	return path.Base(d.PkgPath)
}

func RunView(data RunData) *View {
	return NewTemplateView(RunViewType, "renderRun", data)
}
