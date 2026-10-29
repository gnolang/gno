package run

import (
	"net/http"
	"path"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// GetRunView renders the run page of the package at u, whose functions
// funcs lists (from its qdoc): the caller has checked that it exists.
func (h *Handler) GetRunView(u *weburl.GnoURL, funcs []*doc.JSONFunc) (int, *components.View) {
	pkgPath := path.Join(h.deps.Domain, u.Path)
	return http.StatusOK, NewPageView(RunData{
		PkgPath: pkgPath,
		Script:  starterScript(pkgPath, funcs),
		Remote:  h.deps.Remote,
		ChainId: h.deps.ChainId,
		Gas:     components.RunGas,
	})
}
