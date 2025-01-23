package ruler

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
)

const v3 = "gno.land/r/gov/dao/v3"
const v4 = "gno.land/r/gov/dao/v4"
const v5 = "gno.land/r/gov/dao/v5"
const v6 = "gno.land/r/gov/dao/v6"

const invalid  = "gno.land/r/invalid/dao"

func TestBridge_Functions(t *testing.T) {

	// invalid package cannot add a new dao in charge
	std.TestSetRealm(std.NewCodeRealm(invalid))
	urequire.PanicsWithMessage(t, "package gno.land/r/gov/dao/v4 cannot be added by gno.land/r/invalid/dao", func() {
		New(v4)
	})


	// dao in charge can add a new dao
	std.TestSetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		New(v4)
	})


	// v4 that is in charge adds v5 in charge
	std.TestSetRealm(std.NewCodeRealm(v4))
	urequire.NotPanics(t, func() {
		New(v5)
	})

	// v3 is not in charge anymore even if it was
	std.TestSetRealm(std.NewCodeRealm(v3))
	urequire.PanicsWithMessage(t, "package gno.land/r/gov/dao/v5 cannot be added by gno.land/r/gov/dao/v3", func() {
		New(v5)
	})

	urequire.Equal(
		t,
	`Current GovDAO in charge: gno.land/r/gov/dao/v5

Previous GovDAO implementations in charge: 
- gno.land/r/gov/dao/v3
- gno.land/r/gov/dao/v4

Previous deleted GovDAOs due to security issues or similar: 
`,
		Render(""),
	)

	// v3 can rollback because it is a previous govDAO
	std.TestSetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		Rollback(v4)
	})

	urequire.Equal(
		t,
	`Current GovDAO in charge: gno.land/r/gov/dao/v4

Previous GovDAO implementations in charge: 
- gno.land/r/gov/dao/v3

Previous deleted GovDAOs due to security issues or similar: 
`,
		Render(""),
	)

	std.TestSetRealm(std.NewCodeRealm(v3))
	urequire.PanicsWithMessage(t, "package gno.land/r/gov/dao/v6 cannot be added by gno.land/r/gov/dao/v3", func() {
		New(v6)
	})

	std.TestSetRealm(std.NewCodeRealm(v4))
	urequire.NotPanics(t, func() {
		New(v6)
	})

	std.TestSetRealm(std.NewCodeRealm(v6))
	urequire.NotPanics(t, func() {
		Remove(v3)
	})

	urequire.Equal(
		t,
	`Current GovDAO in charge: gno.land/r/gov/dao/v6

Previous GovDAO implementations in charge: 
- gno.land/r/gov/dao/v4

Previous deleted GovDAOs due to security issues or similar: 
- gno.land/r/gov/dao/v3
`,
		Render(""),
	)
}
