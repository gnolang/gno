// Package ifaceimporter imports ifaceattack, so that package's methods act for
// it whether named or reached through its interface.
package ifaceimporter

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/p/demo/tests/launderpkg/v0"
)

type Runner interface{ Run(*launderpkg.Immutable) }

var gObj = &launderpkg.Immutable{Field: "victim-orig"}

func Read() string  { return gObj.Field }
func Use(r Runner)  { r.Run(gObj) }
func UseNamed()     { ifaceattack.Evil(0).Run(gObj) }
