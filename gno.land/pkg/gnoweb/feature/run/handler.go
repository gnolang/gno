package run

import (
	"net/http"
	"path"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func (h *Handler) GetRunView(u *weburl.GnoURL) (int, *components.View) {
	return http.StatusOK, NewPageView(RunData{
		PkgPath: path.Join(h.deps.Domain, u.Path),
		Domain:  h.deps.Domain,
		Remote:  h.deps.Remote,
		ChainId: h.deps.ChainId,
	})
}
