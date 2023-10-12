package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gotuna/gotuna/test/assert"
)

func TestRoutes(t *testing.T) {
	ok := http.StatusOK
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
	}
	if wd, err := os.Getwd(); err == nil {
		if strings.HasSuffix(wd, "cmd/gnoweb") {
			os.Chdir("../..")
		}
	} else {
		panic("os.Getwd() -> err: " + err.Error())
	}

	config := integration.DefaultTestingNodeConfig(t, gnoland.MustGuessGnoRootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	defer node.Stop()

	// XXX: this is ugly :(
	flags.remoteAddr = node.Config().RPC.ListenAddress

	app := makeApp()

	for _, r := range routes {
		t.Run(fmt.Sprintf("test route %s", r.route), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, r.route, nil)
			response := httptest.NewRecorder()
			app.Router.ServeHTTP(response, request)
			assert.Equal(t, r.status, response.Code)
			assert.Equal(t, strings.Contains(response.Body.String(), r.substring), true)
			println(response.Body.String())
		})
	}
}
