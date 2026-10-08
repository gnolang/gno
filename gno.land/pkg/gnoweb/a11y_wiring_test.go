package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/gnolang/gno/tm2/pkg/log"
)

// TestRailA11yWiring renders every view that draws the sidebar rail and checks
// the id-based wiring the templates rely on, which nothing else would catch:
// renaming an id on one side keeps every page rendering fine.
//   - every aria-labelledby value names an id on the page;
//   - the skip link points at exactly one element, and the first focusable
//     element after it is not in the rail, or skipping would land in the rail;
//   - the rail toggle checkbox stays .u-sr-only: display:none would drop it
//     from the tab order again.
func TestRailA11yWiring(t *testing.T) {
	cfg := NewDefaultAppConfig()
	cfg.NodeRemote = sharedNodeRemote(t)
	router, err := NewRouter(log.NewTestingLogger(t), cfg)
	require.NoError(t, err)

	routes := []struct {
		route   string
		hasRail bool
	}{
		{"/", true},                                        // realm
		{"/r/gnoland/blog$help", true},                     // $help, rail after the header
		{"/r/gnoland/blog$source", true},                   // package overview
		{"/r/gnoland/blog$source&file=admin.gno", true},    // source file
		{"/r/gnoland/blog$state", true},                    // package state
		{"/r/gnoland/blog/", false},                        // directory, no rail
		{"/r/tests/vm/deep/very/deep", true},               // empty TOC: rail emitted (empty <ul>, still shown)
		{"/r/gnoland/blog$source&file=nonexistent", false}, // error page
	}

	for _, r := range routes {
		t.Run(r.route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, r.route, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			doc, err := html.Parse(rec.Body)
			require.NoError(t, err)

			var (
				nodes  []*html.Node // document order
				ids    = map[string]int{}
				toggle *html.Node
			)
			for n := range doc.Descendants() {
				if n.Type != html.ElementNode {
					continue
				}
				nodes = append(nodes, n)
				id := attr(n, "id")
				if id != "" {
					ids[id]++
				}
				if id == "toc-expend" {
					toggle = n
				}
			}

			for _, n := range nodes {
				for ref := range strings.FieldsSeq(attr(n, "aria-labelledby")) {
					assert.Equal(t, 1, ids[ref], "<%s aria-labelledby=%q> must name exactly one id on the page", n.Data, ref)
				}
			}

			if r.hasRail {
				require.Equal(t, 1, ids["toc-expend"], "rail toggle #toc-expend must exist exactly once")
				assert.True(t, hasClass(toggle, "u-sr-only"), "rail toggle must stay visually hidden, not display:none")
			}

			skip := slices.IndexFunc(nodes, func(n *html.Node) bool { return hasClass(n, "u-skip-link") })
			require.NotEqual(t, -1, skip, "skip link missing")
			href := attr(nodes[skip], "href")
			require.True(t, strings.HasPrefix(href, "#"), "skip link href %q is not a fragment", href)
			target := strings.TrimPrefix(href, "#")
			require.Equal(t, 1, ids[target], "skip link target #%s must exist exactly once", target)

			at := slices.IndexFunc(nodes, func(n *html.Node) bool { return attr(n, "id") == target })
			assert.Equal(t, "-1", attr(nodes[at], "tabindex"), "skip target needs tabindex=-1 to take focus")
			next := slices.IndexFunc(nodes[at+1:], isFocusable)
			if next == -1 {
				return
			}
			first := nodes[at+1+next]
			for p := first; p != nil; p = p.Parent {
				assert.False(t, hasClass(p, "b-sidebar"), "first Tab after the skip target lands in the rail (<%s>)", first.Data)
			}
		})
	}
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, c string) bool {
	return n.Type == html.ElementNode && slices.Contains(strings.Fields(attr(n, "class")), c)
}

// isFocusable approximates the sequential focus order: what a Tab can reach.
func isFocusable(n *html.Node) bool {
	if attr(n, "tabindex") == "-1" {
		return false
	}
	if attr(n, "tabindex") != "" {
		return true
	}
	switch n.Data {
	case "a":
		return attr(n, "href") != ""
	case "input":
		return attr(n, "type") != "hidden"
	case "button", "select", "textarea", "summary":
		return true
	}
	return false
}
