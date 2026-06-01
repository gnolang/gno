package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsHandlerRoutes exercises the /docs subtree end-to-end through the
// full router so we know the embed, the resolver, the link rewriter, and
// the layout all wire together. It is the minimum guard against /docs
// breaking silently when something nearby changes.
func TestDocsHandlerRoutes(t *testing.T) {
	t.Parallel()

	logger := log.NewTestingLogger(t)
	rootdir := gnoenv.RootDir()
	genesis := integration.LoadDefaultGenesisTXsFile(t, "tendermint_test", rootdir)
	config, _ := integration.TestingNodeConfig(t, rootdir, genesis...)
	node, remoteAddr := integration.TestingInMemoryNode(t, logger, config)
	t.Cleanup(func() { node.Stop() })

	cfg := NewDefaultAppConfig()
	cfg.NodeRemote = remoteAddr

	router, err := NewRouter(logger, cfg)
	require.NoError(t, err)

	cases := []struct {
		name        string
		route       string
		wantStatus  int
		wantSnippet string // substring expected in the rendered body
	}{
		{
			name:        "index renders README",
			route:       "/docs",
			wantStatus:  http.StatusOK,
			wantSnippet: "Welcome to the official documentation",
		},
		{
			name:        "trailing slash index",
			route:       "/docs/",
			wantStatus:  http.StatusOK,
			wantSnippet: "Welcome to the official documentation",
		},
		{
			name:        "sub-page renders",
			route:       "/docs/builders/getting-started",
			wantStatus:  http.StatusOK,
			wantSnippet: "Getting started",
		},
		{
			name:        "links rewritten to /docs/clean URLs",
			route:       "/docs",
			wantStatus:  http.StatusOK,
			wantSnippet: `href="/docs/builders/getting-started"`,
		},
		{
			name:        "unknown page is 404",
			route:       "/docs/does-not-exist",
			wantStatus:  http.StatusNotFound,
			wantSnippet: "",
		},
		{
			// getting-started.md uses :::tip and :::warning admonitions;
			// confirm the transform reaches the renderer and produces the
			// gno-alert-* markup emitted by markdown/ext_alert.go.
			name:        "admonitions render as gno-alert elements",
			route:       "/docs/builders/getting-started",
			wantStatus:  http.StatusOK,
			wantSnippet: "gno-alert-tip",
		},
		{
			name:        "sidebar section heading present",
			route:       "/docs",
			wantStatus:  http.StatusOK,
			wantSnippet: ">Build on Gno.land<",
		},
		{
			name:        "sidebar item points at clean URL",
			route:       "/docs",
			wantStatus:  http.StatusOK,
			wantSnippet: `href="/docs/builders/getting-started"`,
		},
		{
			name:        "current page item marked active",
			route:       "/docs/builders/getting-started",
			wantStatus:  http.StatusOK,
			wantSnippet: "b-docs-nav-item is-active",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tc.route, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code, "status mismatch")
			if tc.wantSnippet != "" {
				assert.True(t,
					strings.Contains(rec.Body.String(), tc.wantSnippet),
					"expected %q in body", tc.wantSnippet,
				)
			}
		})
	}
}

func TestRewriteDocsLinks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		currentRel string
		in         string
		want       string
	}{
		{
			name:       "relative .md link from README",
			currentRel: "README.md",
			in:         `see [start](builders/getting-started.md)`,
			want:       `see [start](/docs/builders/getting-started)`,
		},
		{
			name:       "parent-relative .md link",
			currentRel: "builders/foo.md",
			in:         `see [resources](../resources/effective-gno.md)`,
			want:       `see [resources](/docs/resources/effective-gno)`,
		},
		{
			name:       "anchor preserved",
			currentRel: "README.md",
			in:         `[s](builders/foo.md#section)`,
			want:       `[s](/docs/builders/foo#section)`,
		},
		{
			name:       "absolute URL untouched",
			currentRel: "README.md",
			in:         `[g](https://github.com/gnolang/gno)`,
			want:       `[g](https://github.com/gnolang/gno)`,
		},
		{
			name:       "fragment-only untouched",
			currentRel: "README.md",
			in:         `[top](#top)`,
			want:       `[top](#top)`,
		},
		{
			name:       "image rewritten and extension preserved",
			currentRel: "README.md",
			in:         `![logo](images/logo.png)`,
			want:       `![logo](/docs/images/logo.png)`,
		},
		{
			name:       "cross-repo link escapes to GitHub",
			currentRel: "resources/gno-data-structures.md",
			in:         `[avl](../../examples/gno.land/p/nt/avl/v0/README.md)`,
			want:       `[avl](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/nt/avl/v0/README.md)`,
		},
		{
			name:       "cross-repo link with anchor",
			currentRel: "resources/gno-interrealm.md",
			in:         `[adr](../../gnovm/adr/interrealm_v2.md#summary)`,
			want:       `[adr](https://github.com/gnolang/gno/blob/master/gnovm/adr/interrealm_v2.md#summary)`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := string(rewriteDocsLinks([]byte(tc.in), tc.currentRel))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTransformAdmonitions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "simple info",
			in:   ":::info\nhello\nworld\n:::\n",
			want: "> [!INFO]\n> hello\n> world\n\n",
		},
		{
			name: "tip with inline title",
			in:   ":::tip Try this\nbody\n:::\n",
			want: "> [!TIP] Try this\n> body\n\n",
		},
		{
			name: "warning with multi-word title",
			in:   ":::warning Safeguard your mnemonic phrase!\nbody\n:::\n",
			want: "> [!WARNING] Safeguard your mnemonic phrase!\n> body\n\n",
		},
		{
			name: "blank line inside body preserved",
			in:   ":::info\nfirst\n\nsecond\n:::\n",
			want: "> [!INFO]\n> first\n>\n> second\n\n",
		},
		{
			name: "passthrough when no admonitions",
			in:   "# Title\nbody\n",
			want: "# Title\nbody\n",
		},
		{
			name: "code fence with ::: inside is untouched",
			in:   "```\n:::info\nstill code\n:::\n```\n",
			want: "```\n:::info\nstill code\n:::\n```\n",
		},
		{
			name: "admonition before and after a code fence",
			in:   ":::tip\nbefore\n:::\n```\nx\n```\n:::warning\nafter\n:::\n",
			want: "> [!TIP]\n> before\n\n```\nx\n```\n> [!WARNING]\n> after\n\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := string(transformAdmonitions([]byte(tc.in)))
			assert.Equal(t, tc.want, got)
		})
	}
}
