package home

import (
	"std"
	"strings"
)

// RenderFn defines the type for the render function of realms.
type RenderFn func(string) string

var (
	proxyRealmPath string
	current        = struct {
		realmPath string
		renderFn  RenderFn
	}{}
)

func init() {
	proxyRealmPath = std.CurrentRealm().PkgPath()
}

// CurrentRealmPath returns the path of the realm that is currently registered.
func CurrentRealmPath() string {
	return current.realmPath
}

// Register registers a render function of a realm.
func Register(fn RenderFn) {
	crossing()

	if fn == nil {
		panic("render function must not be nil")
	}

	if std.PreviousRealm().IsUser() {
		panic("register must be called by a realm")
	}

	callerPath := std.PreviousRealm().PkgPath()
	if !strings.HasPrefix(callerPath, proxyRealmPath+"/") {
		panic("caller realm path must start with " + proxyRealmPath)
	}

	current.renderFn = fn
	current.realmPath = callerPath
}

// URL returns a URL that links to the proxy realm.
func URL(renderPath string) string {
	url := "http://" + proxyRealmPath
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
