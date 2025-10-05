package core

import (
	"chain/runtime"
	"errors"
	"regexp"

	"gno.land/p/aib/ibc/app"
	"gno.land/p/nt/ufmt"
)

// isAlphaNumeric defines a regular expression for matching against
// alpha-numeric values.
var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// RegisterApp registers an IBCApp, the portID is determined using the package
// path of the caller, hence it must an other realm.
func RegisterApp(cur realm, portID string, app app.IBCApp) {
	caller := runtime.PreviousRealm()
	if caller.IsUser() {
		panic("RegisterApp can only be called from an other realm")
	}
	if !isAlphaNumeric(portID) {
		panic(errors.New("portID can only contain alphanumeric characters"))
	}
	if _, ok := store.routes[portID]; ok {
		panic(ufmt.Errorf("portID %s has already been registered", portID))
	}
	store.routes[portID] = app
}
