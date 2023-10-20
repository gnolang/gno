package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
	if wd, err := os.Getwd(); err == nil {
		if strings.HasSuffix(wd, "cmd/gnoweb") {
			os.Chdir("../..")
		}
	} else {
		panic("os.Getwd() -> err: " + err.Error())
	}

	// configure default values
	flags.RemoteAddr = "127.0.0.1:26657"
	flags.HelpRemote = "127.0.0.1:26657"
	flags.HelpChainID = "dev"
	flags.CaptchaSite = ""
	flags.ViewsDir = "./cmd/gnoweb/views"
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

	t.Run("with", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				flags.WithAnalytics = true
				app := makeApp()
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Contains(t, response.Body.String(), "simpleanalytics")
			})
		}
	})
	t.Run("without", func(t *testing.T) {
		for _, route := range routes {
			t.Run(route, func(t *testing.T) {
				flags.WithAnalytics = false
				app := makeApp()
				request := httptest.NewRequest(http.MethodGet, route, nil)
				response := httptest.NewRecorder()
				app.Router.ServeHTTP(response, request)
				assert.Equal(t, strings.Contains(response.Body.String(), "simpleanalytics"), false)
			})
		}
	})
}
