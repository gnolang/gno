package connect

import (
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// ServeHTTP renders /wallets from the embedded registry. Server-rendered from
// the same data the chooser reads, so it works offline, on any chain, and
// cannot drift from what the chooser offers.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := components.IndexData{
		HeadData: components.HeadData{
			Title:             "gno.land — compatible wallets",
			Description:       "Wallets that work with gno.land, and where to install them.",
			AssetsPath:        h.deps.Meta.AssetsPath,
			ChromaPath:        h.deps.Meta.ChromaPath,
			ChainId:           h.deps.Meta.ChainId,
			Remote:            h.deps.Meta.Remote,
			BuildTime:         h.deps.Meta.BuildTime,
			AnalyticsHostname: h.deps.Meta.AnalyticsHostname,
		},
		FooterData: components.FooterData{
			Analytics: components.AnalyticsData{Enabled: h.deps.Meta.Analytics},
		},
		Mode:     components.ViewModeHome,
		Banner:   h.deps.Meta.Banner,
		BodyView: NewWalletsView(WalletsData{Groups: GroupWallets(components.Wallets())}),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := components.IndexLayout(data).Render(w); err != nil {
		h.deps.Logger.Error("failed to render wallets page", "error", err)
	}
}
