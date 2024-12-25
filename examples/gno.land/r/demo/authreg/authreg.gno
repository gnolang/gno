package authreg

import (
	"errors"
	"path"
	"std"

	"gno.land/p/demo/auth"
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
	return path.Join("/", provider, authFn(autok))
}
