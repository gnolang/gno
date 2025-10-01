package core

import (
	"gno.land/p/aib/ibc/app"
	"gno.land/p/nt/ufmt"
)

func RegisterApp(cur realm, portID string, app app.IBCApp) {
	// TODO assert this call cannot come from a user
	if _, ok := store.routes[portID]; ok {
		panic(ufmt.Errorf("route %s has already been registered", portID))

	}
	store.routes[portID] = app
}
