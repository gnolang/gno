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
		route  string
		status int
		ftest  func(string) bool
	}{
		{"/", ok, func(s string) bool { return strings.Contains(s, "Welcome") }},
		{"/about", ok, func(s string) bool { return strings.Contains(s, "blockchain") }},
		{"/r/gnoland/blog", ok, nil},
		{"/r/gnoland/blog?help", ok, func(s string) bool { return strings.Contains(s, "exposed") }},
		{"/r/gnoland/blog/", ok, func(s string) bool { return strings.Contains(s, "admin.gno") }},
		{"/r/gnoland/blog/admin.gno", ok, func(s string) bool { return strings.Contains(s, "func ") }},
		{"/r/demo/users:administrator", ok, func(s string) bool { return strings.Contains(s, "address") }},
		{"/r/demo/users", ok, func(s string) bool { return strings.Contains(s, "manfred") }},
		{"/r/demo/users/types.gno", ok, func(s string) bool { return strings.Contains(s, "type ") }},
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
			if r.ftest != nil {
				assert.Equal(t, r.ftest(response.Body.String()), true)
			}
		})
	}
}
