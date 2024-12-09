package gnoweb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gotuna/gotuna/test/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutes(t *testing.T) {
	const (
		ok       = http.StatusOK
		found    = http.StatusFound
		notFound = http.StatusNotFound
	)
	routes := []struct {
		route     string
		status    int
		substring string
	}{
		{"/", ok, "Welcome"}, // assert / gives 200 (OK). assert / contains "Welcome".
		{"/about", ok, "blockchain"},
		{"/r/gnoland/blog", ok, ""}, // whatever content
		{"/r/gnoland/blog$help", ok, "AdminSetAdminAddr"},
		{"/r/gnoland/blog/", ok, "admin.gno"},
		{"/r/gnoland/blog/admin.gno", ok, ">func<"},
		{"/r/gnoland/blog$help&func=Render", ok, "Render(path)"},
		{"/r/gnoland/blog$help&func=Render&path=foo/bar", ok, `value="foo/bar"`},
		// {"/r/gnoland/blog$help&func=NonExisting", ok, "NonExisting not found"}, // XXX(TODO)
		{"/r/demo/users:administrator", ok, "address"},
		{"/r/demo/users", ok, "moul"},
		{"/r/demo/users/users.gno", ok, "// State"},
		{"/r/demo/deep/very/deep", ok, "it works!"},
		{"/r/demo/deep/very/deep?arg1=val1&arg2=val2", ok, "hi ?arg1=val1&amp;arg2=val2"},
		{"/r/demo/deep/very/deep:bob", ok, "hi bob"},
		{"/r/demo/deep/very/deep:bob?arg1=val1&arg2=val2", ok, "hi bob?arg1=val1&amp;arg2=val2"},
		{"/r/demo/deep/very/deep$help", ok, "Render"},
		{"/r/demo/deep/very/deep/", ok, "render.gno"},
		{"/r/demo/deep/very/deep/render.gno", ok, ">package<"},
		{"/contribute", ok, "Game of Realms"},
		{"/game-of-realms", found, "/contribute"},
		{"/gor", found, "/contribute"},
		{"/blog", found, "/r/gnoland/blog"},
		{"/404/not/found/", notFound, ""},
		{"/아스키문자가아닌경로", notFound, ""},
		{"/%ED%85%8C%EC%8A%A4%ED%8A%B8", notFound, ""},
		{"/グノー", notFound, ""},
		{"/⚛️", notFound, ""},
		{"/p/demo/flow/LICENSE", ok, "BSD 3-Clause"},
	}

	rootdir := gnoenv.RootDir()
	genesis := integration.LoadDefaultGenesisTXsFile(t, "tendermint_test", rootdir)
	config, _ := integration.TestingNodeConfig(t, rootdir, genesis...)
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewTestingLogger(t), config)
	defer node.Stop()

	cfg := NewDefaultAppConfig()
	cfg.NodeRemote = remoteAddr

	logger := log.NewTestingLogger(t)

	// set the `remoteAddr` of the client to the listening address of the
	// node, which is randomly assigned.
	router, err := MakeRouterApp(logger, cfg)
	require.NoError(t, err)

	for _, r := range routes {
		t.Run(fmt.Sprintf("test route %s", r.route), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, r.route, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, r.status, response.Code)
			assert.Contains(t, response.Body.String(), r.substring)
		})
	}
}

func TestAnalytics(t *testing.T) {
	routes := []string{
		// special realms
		"/", // home
		"/about",
		"/start",

		// redirects
		"/game-of-realms",
		"/getting-started",
		"/blog",
		"/boards",

		// realm, source, help page
		"/r/gnoland/blog",
		"/r/gnoland/blog/admin.gno",
		"/r/demo/users:administrator",
		"/r/gnoland/blog$help",

		// special pages
		"/404-not-found",
	}

	rootdir := gnoenv.RootDir()
	genesis := integration.LoadDefaultGenesisTXsFile(t, "tendermint_test", rootdir)
	config, _ := integration.TestingNodeConfig(t, rootdir, genesis...)
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewTestingLogger(t), config)
	defer node.Stop()

	cfg := NewDefaultAppConfig()
	cfg.NodeRemote = remoteAddr

	logger := log.NewTestingLogger(t)

	t.Run("with", func(t *testing.T) {
		cfg.Analytics = true
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				router, err := MakeRouterApp(logger, cfg)
				require.NoError(t, err)

				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				assert.Contains(t, response.Body.String(), "sa.gno.services")
			})
		}
	})
	t.Run("without", func(t *testing.T) {
		cfg.Analytics = false
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				router, err := MakeRouterApp(logger, cfg)
				require.NoError(t, err)

				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				assert.Equal(t, strings.Contains(response.Body.String(), "sa.gno.services"), false)
			})
		}
	})
}
