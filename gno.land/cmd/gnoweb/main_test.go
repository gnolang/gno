package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
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
		{"/r/demo/users", ok, "manfred"},
		{"/r/demo/users/types.gno", ok, "type "},
		{"/r/demo/deep/very/deep", ok, "it works!"},
		{"/r/demo/deep/very/deep:bob", ok, "hi bob"},
		{"/r/demo/deep/very/deep?help", ok, "exposed"},
		{"/r/demo/deep/very/deep/", ok, "render.gno"},
		{"/r/demo/deep/very/deep/render.gno", ok, "func Render("},
		{"/game-of-realms", ok, "/r/gnoland/pages:p/gor"},
		{"/gor", found, "/game-of-realms"},
		{"/blog", found, "/r/gnoland/blog"},
		{"/404-not-found", notFound, "/404-not-found"},
	}

	config, _ := integration.TestingNodeConfig(t, gnoland.MustGuessGnoRootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	defer node.Stop()

	// set the `remoteAddr` of the client to the listening address of the
	// node, which is randomly assigned.
	flags.RemoteAddr = remoteAddr
	flags.HelpChainID = "dev"
	flags.CaptchaSite = ""
	flags.ViewsDir = "../../cmd/gnoweb/views"
	flags.WithAnalytics = false
	app := makeApp()

	for _, r := range routes {
		t.Run(fmt.Sprintf("test route %s", r.route), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, r.route, nil)
			response := httptest.NewRecorder()
			app.Router.ServeHTTP(response, request)
			assert.Equal(t, r.status, response.Code)
			assert.Contains(t, response.Body.String(), r.substring)
			// println(response.Body.String())
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

	config, _ := integration.TestingNodeConfig(t, gnoland.MustGuessGnoRootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	defer node.Stop()

	flags.ViewsDir = "../../cmd/gnoweb/views"
	t.Run("with", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				flags.RemoteAddr = remoteAddr
				flags.WithAnalytics = true
				app := makeApp()
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
				flags.RemoteAddr = remoteAddr
				flags.WithAnalytics = false
				app := makeApp()
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Equal(t, strings.Contains(response.Body.String(), "sa.gno.services"), false)
			})
		}
	})
}
