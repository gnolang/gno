package authreg

import (
	"errors"
	"path"
	"std"
	"strings"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/ufmt"
)

var fns = make(map[string]auth.AuthenticateFn)

// XXX: we could add a slug there
func Register(authenticate auth.AuthenticateFn) {
	caller := std.PrevRealm().Addr()
	fns[caller.String()] = authenticate
}

func Authenticate(autok auth.Token) string {
	provider := autok.Source().Addr().String()
	authFn, ok := fns[provider]
	if !ok {
		panic(errors.New("unknown auth provider"))
	}

	subPath := path.Clean(authFn(autok))
	if strings.HasPrefix(subPath, "..") {
		panic(ufmt.Errorf("invalid sub-path %q", subPath))
	}

	return path.Join("/", provider, subPath)
}
