package core

import (
	"chain/runtime"

	"gno.land/p/aib/ibc/app"
	"gno.land/p/nt/ufmt"
)

// RegisterApp registers an IBCApp, the portID is determined using the package
// path of the caller, hence it must an other realm.
func RegisterApp(cur realm, app app.IBCApp) {
	caller := runtime.PreviousRealm()
	if caller.IsUser() {
		panic("RegisterApp can only be called from an other realm")
	}
	portID := caller.PkgPath()
	if _, ok := store.routes[portID]; ok {
		panic(ufmt.Errorf("route %s has already been registered", portID))
	}
	store.routes[portID] = app
}
