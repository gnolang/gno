package main

import (
	"fmt"
	"github.com/gotuna/gotuna/test/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
	}
	if wd, err := os.Getwd(); err == nil {
		if strings.HasSuffix(wd, "gnoland/website") {
			os.Chdir("../..")
		}
	} else {
		panic("os.Getwd() -> err: " + err.Error())
	}
	app := makeApp()
	for _, r := range routes {
		t.Run(fmt.Sprintf("test route %s", r.route), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, r.route, nil)
			response := httptest.NewRecorder()
			app.Router.ServeHTTP(response, request)
			assert.Equal(t, r.status, response.Code)
			assert.Equal(t, strings.Contains(response.Body.String(), r.substring), true)
		})
	}
}
