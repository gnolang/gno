package browser

import (
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
)

func redirectWebPath(path string) string {
	if alias, ok := gnoweb.Aliases[path]; ok {
		return alias
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
	path = filepath.Clean(path)

	return path
}
