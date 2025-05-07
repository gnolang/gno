package browser

import (
	gopath "path"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

func redirectWebPath(path string) string {
	if alias, ok := gnoweb.DefaultAliases[path]; ok {
		if alias.Kind == gnoweb.GnowebPath { // Ignore static files
			return alias.Value
		}
	}

	if redirect, ok := gnoweb.Redirects[path]; ok {
		return redirect
	}

	return path
}

func cleanupRealmPath(prefix, realm string) string {
	// Trim prefix
	path := strings.TrimPrefix(realm, prefix)
	// redirect if any well known path
	path = redirectWebPath(path)
	// trim any slash
	path = strings.TrimPrefix(path, "/")
	// clean up path
	path = gopath.Clean(path)

	return path
}
