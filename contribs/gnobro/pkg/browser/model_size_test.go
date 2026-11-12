package browser

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
)

// renderRPC answers every query with the same realm page.
type renderRPC struct {
	rpcclient.Client
}

func (renderRPC) ABCIQuery(context.Context, string, []byte) (*ctypes.ResultABCIQuery, error) {
	page := "# Realm\n\nA paragraph.\n\n    a code block\n"
	return &ctypes.ResultABCIQuery{Response: abci.ResponseQuery{
		ResponseBase: abci.ResponseBase{Data: []byte(page)},
	}}, nil
}

// renderAt returns the frame and the realm page gnobro renders for a client
// that declares a width x height terminal, at connection or, when resized is
// set, as a window change after an 80x24 start.
func renderAt(t *testing.T, banner, resized bool, width, height int) (frame, page string) {
	t.Helper()

	cfg := DefaultConfig()
	if banner {
		cfg.Banner = NewModelBanner(0, []string{"gno"})
	}
	var tm tea.Model = New(cfg, &gnoclient.Client{RPCClient: renderRPC{}})
	if resized {
		tm, _ = tm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	}
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: width, Height: height})
	tm, _ = tm.Update(tickBannerMsg{}) // draws the banner's first frame

	view, err := tm.(model).fetchRenderView(cfg.URLDefaultValue)
	if err != nil {
		t.Fatal(err)
	}
	return tm.View(), string(view)
}

// Over SSH a client declares its terminal size, and every view pads to it. A
// size past the cap renders as the cap, and one cell under it does not.
func TestDeclaredSizeCap(t *testing.T) {
	const huge = 1 << 15 // past any terminal, small enough to render unclamped

	type size struct{ w, h int }
	for _, banner := range []bool{false, true} {
		for _, resized := range []bool{false, true} {
			for _, tc := range []struct{ under, capped, huge size }{
				{size{2047, 40}, size{2048, 40}, size{huge, 40}},
				{size{80, 1023}, size{80, 1024}, size{80, huge}},
			} {
				fu, pu := renderAt(t, banner, resized, tc.under.w, tc.under.h)
				fc, pc := renderAt(t, banner, resized, tc.capped.w, tc.capped.h)
				fh, ph := renderAt(t, banner, resized, tc.huge.w, tc.huge.h)
				if fh != fc || ph != pc {
					t.Errorf("banner=%v resized=%v: %v renders unlike the cap %v", banner, resized, tc.huge, tc.capped)
				}
				if fu == fc && pu == pc {
					t.Errorf("banner=%v resized=%v: %v renders as the cap %v", banner, resized, tc.under, tc.capped)
				}
			}
		}
	}
}
