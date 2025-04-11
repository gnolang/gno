package home

import (
	"chain/runtime"
	"strings"
)

// RenderFn defines the type for the render function of realms.
type RenderFn func(string) string

var current = struct {
	realmPath string
	renderFn  RenderFn
}{}

// CurrentRealmPath returns the path of the realm that is currently registered.
func CurrentRealmPath() string {
	return current.realmPath
}

// Register registers a render function of a realm.
func Register(fn RenderFn) {
	if fn == nil {
		panic("render function must not be nil")
	}

	proxyPath := runtime.CurrentRealm().PkgPath()
	callerPath := runtime.PreviousRealm().PkgPath()
	if !strings.HasPrefix(callerPath, proxyPath+"/") {
		panic("caller realm path must start with " + proxyPath)
	}

	current.renderFn = fn
	current.realmPath = callerPath
}

// URL returns a URL that links to the proxy realm.
func URL(renderPath string) string {
	url := "http://" + runtime.CurrentRealm().PkgPath()
	if renderPath != "" {
		url += ":" + renderPath
	}
	return url
}

// Render renders the rendered Markdown of the realm that is currently registered.
func Render(path string) string {
	if current.renderFn == nil {
		panic("no realm has been registered")
	}
	return current.renderFn(path)
}
