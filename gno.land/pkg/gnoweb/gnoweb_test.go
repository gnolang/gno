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
		{"/r/gnoland/blog?help", ok, "exposed"},
		{"/r/gnoland/blog/", ok, "admin.gno"},
		{"/r/gnoland/blog/admin.gno", ok, "func "},
		{"/r/demo/users:administrator", ok, "address"},
		{"/r/demo/users", ok, "moul"},
		{"/r/demo/users/users.gno", ok, "// State"},
		{"/r/demo/deep/very/deep", ok, "it works!"},
		{"/r/demo/deep/very/deep:bob", ok, "hi bob"},
		{"/r/demo/deep/very/deep?help", ok, "exposed"},
		{"/r/demo/deep/very/deep/", ok, "render.gno"},
		{"/r/demo/deep/very/deep/render.gno", ok, "func Render("},
		{"/contribute", ok, "Game of Realms"},
		{"/game-of-realms", found, "/contribute"},
		{"/gor", found, "/contribute"},
		{"/blog", found, "/r/gnoland/blog"},
		{"/404-not-found", notFound, "/404-not-found"},
		{"/아스키문자가아닌경로", notFound, "/아스키문자가아닌경로"},
		{"/%ED%85%8C%EC%8A%A4%ED%8A%B8", notFound, "/테스트"},
		{"/グノー", notFound, "/グノー"},
		{"/⚛️", notFound, "/⚛️"},
		{"/p/demo/flow/LICENSE", ok, "BSD 3-Clause"},
	}

	rootdir := gnoenv.RootDir()
	genesis := integration.LoadDefaultGenesisTXsFile(t, "tendermint_test", rootdir)
	config, _ := integration.TestingNodeConfig(t, rootdir, genesis...)
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewTestingLogger(t), config)
	defer node.Stop()

	cfg := NewDefaultConfig()

	logger := log.NewTestingLogger(t)

	// set the `remoteAddr` of the client to the listening address of the
	// node, which is randomly assigned.
	cfg.RemoteAddr = remoteAddr
	app := MakeApp(logger, cfg)

	for _, r := range routes {
		t.Run(fmt.Sprintf("test route %s", r.route), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, r.route, nil)
			response := httptest.NewRecorder()
			app.Router.ServeHTTP(response, request)
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
		"/r/gnoland/blog?help",

		// special pages
		"/404-not-found",
	}

	rootdir := gnoenv.RootDir()
	genesis := integration.LoadDefaultGenesisTXsFile(t, "tendermint_test", rootdir)
	config, _ := integration.TestingNodeConfig(t, rootdir, genesis...)
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewTestingLogger(t), config)
	defer node.Stop()

	cfg := NewDefaultConfig()
	cfg.RemoteAddr = remoteAddr

	logger := log.NewTestingLogger(t)

	t.Run("with", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				ccfg := cfg // clone config
				ccfg.WithAnalytics = true
				app := MakeApp(logger, ccfg)
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Contains(t, response.Body.String(), "sa.gno.services")
			})
		}
	})
	t.Run("without", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				ccfg := cfg // clone config
				ccfg.WithAnalytics = false
				app := MakeApp(logger, ccfg)
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Equal(t, strings.Contains(response.Body.String(), "sa.gno.services"), false)
			})
		}
	})
}
