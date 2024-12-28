package authreg

import (
	"errors"
	"std"

	"gno.land/p/demo/auth"
)

var fns = make(map[string]auth.AuthenticateFn)

// XXX: we could add a slug there
func Register(authenticate auth.AuthenticateFn) {
	caller := std.PrevRealm()
	if caller.IsUser() {
		panic("can't register from user realm")
	}
	callerStr := caller.Addr().String()

	if authenticate == nil {
		if _, ok := fns[callerStr]; !ok {
			panic(errors.New("not registered"))
		}

		delete(fns, callerStr)
		std.Emit(
			unregisterEvent,
			"caller", callerStr,
		)
		return
	}

	fns[callerStr] = authenticate
	std.Emit(
		registerEvent,
		"caller", callerStr,
	)
}

func Authenticate(autok auth.Token) string {
	provider := autok.Source().Addr().String()
	authFn, ok := fns[provider]
	if !ok {
		panic(errors.New("unknown auth provider"))
	}

	// XXX: maybe emit event for audit purposes

	return auth.NamespacedEntityID(provider, authFn(autok))
}

const (
	registerEvent   = "register"
	unregisterEvent = "unregister"
)
