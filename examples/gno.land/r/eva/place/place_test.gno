package place

import (
	"encoding/base64"
	"std"
	"testing"

	"gno.land/p/demo/ufmt"
)

var (
	testValidColor = []string{
		"rgb(255,0,0)",
		"rgb(0,0,0)",
		"rgb(255,255,255)",
	}

	testInvalidColor = []string{
		"rgb(255555,0,0)",
		"rgb(0,0)",
		"rgb(,0,0)",
		"rgb(1,0;0)",
		"rgb(1k,0,0)",
		"rgb(0,256,0)",
	}

	testSubmitInvalidCoordinates = []struct {
		x uint64
		y uint64
	}{
		{1000, 1},
		{1, 40},
	}

	testSubmitInvalidColor = []string{
		"This color is definitly not a valid one",
		"",
		"red",
		"Blue",
	}

	testValidAdress = []struct {
		address string
		x       uint64
		y       uint64
	}{
		{"someAdress/r/eva/place:?color=G&x=9&y=2", 9, 2},
		{"?x=19&y=1", 19, 1},
	}

	testInvalidAdress = []string{
		"someAdress/r/eva/place:?color=G&x=9a&y=2",
		"",
	}

	randomTestWallet  = "g1uz7u2aeff4z8haftcsay2zp2rcc5knkru3y4im"
	randomTestWallet2 = "g1uz7u2aeffty8hafttest2zp2rcc5knkru3y4im"
	bobAdminWannabe   = "g1uz7u2aeffty8haft73985782rtt5knkruadmin"
	adminList         = []string{
		louWallet,
		chaeWallet,
		ghhvWallet,
	}
)

func Test_generateSquareSvg_valid(t *testing.T) {
	value, err := generateSquareSvg("rgb(255,0,0)")

	if err != nil {
		t.Fatal("invalid color")
	}

	decodedData, err := base64.StdEncoding.DecodeString(value)

	correctValue := `<svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox=""><rect x="0" y="0" width="30" height="30" rx="0" ry="0" fill="rgb(255,0,0)" /></svg>`

	if err != nil {
		t.Fatal("Cannot decode the svg")
	}
	// Verify the svg is correct
	if string(decodedData) != correctValue {
		t.Fatalf("The svg created is invalid")
	}
}

func Test_generateSquareSvg_invalid(t *testing.T) {
	for _, testInput := range testInvalidColor {
		t.Run(testInput, func(t *testing.T) {
			_, err := generateSquareSvg(testInput)

			if err == nil {
				t.Fatal("color should be invalid")
			}
		})
	}
}

func Test_checkColorValid_true(t *testing.T) {
	for _, testInput := range testValidColor {
		t.Run(testInput, func(t *testing.T) {
			value := checkColorValid(testInput)

			// Verify the svg is correct
			if value == false {
				t.Fatalf("The color should be valid")
			}
		})
	}
}

func Test_checkColorValid_false(t *testing.T) {
	for _, testInput := range testInvalidColor {
		t.Run(testInput, func(t *testing.T) {
			value := checkColorValid(testInput)

			// Verify the svg is incorrect
			if value == true {
				t.Fatalf("The color should be invalid")
			}
		})
	}
}

func Test_submitInvalidCoordinates(t *testing.T) {
	for _, testInput := range testSubmitInvalidCoordinates {
		testName := ufmt.Sprintf("(%d,%d)", testInput.x, testInput.y)
		t.Run(testName, func(t *testing.T) {
			abort := revive(func() {
				Submit(cross, testInput.x, testInput.y, "R")
			})

			if abort == nil {
				t.Fatal("Invalid Submit didn't panic")
			}
		})
	}
}

func Test_submitInvalidColor(t *testing.T) {
	for _, testInput := range testSubmitInvalidColor {
		t.Run(testInput, func(t *testing.T) {
			abort := revive(func() {
				Submit(cross, 0, 0, testInput)
			})

			if abort == nil {
				t.Fatal("Invalid Submit didn't panic")
			}
		})
	}
}

func Test_submitValid(t *testing.T) {
	tableColor[39][1] = Color{"rgb(250,250,250)", "W"}

	Submit(cross, 1, 39, "R")

	if tableColor[39][1].initial != "R" {
		t.Fatal("Submit not taken into account")
	}
}

func Test_parseXYValid(t *testing.T) {
	for _, testInput := range testValidAdress {
		t.Run(testInput.address, func(t *testing.T) {
			x, y, err := parseXY(testInput.address)

			if err != nil || x != testInput.x || y != testInput.y {
				t.Fatal("coordinates should be valid")
			}
		})
	}
}

func Test_parseXYInvalid(t *testing.T) {
	for _, testInput := range testInvalidAdress {
		t.Run(testInput, func(t *testing.T) {
			_, _, err := parseXY(testInput)

			if err == nil {
				t.Fatal("coordinates should be invalid")
			}
		})
	}
}

func Test_Register(t *testing.T) {
	owner := std.Address(randomTestWallet)
	testing.SetOriginCaller(owner)

	abort := revive(func() {
		Register(cross)
	})

	if abort != nil {
		t.Fatal("Register panicked for first register")
	}

	abort = revive(func() {
		Register(cross)
	})

	if abort == nil {
		t.Fatal("Register didn't panick for second register")
	}
}

func Test_DeleteYourself(t *testing.T) {
	owner := std.Address(randomTestWallet)
	testing.SetOriginCaller(owner)

	revive(func() {
		Register(cross)
	})

	abort := revive(func() {
		DeleteYourself(cross)
	})

	if abort != nil {
		t.Fatal("DeleteYourself panicked for first removal")
	}

	abort = revive(func() {
		DeleteYourself(cross)
	})

	if abort == nil {
		t.Fatal("DeleteYourself didn't panick for second removal")
	}
}

func Test_DeleteAllAdmins(t *testing.T) {
	for idx, admin := range adminList {
		owner := std.Address(admin)
		testing.SetOriginCaller(owner)

		abort := revive(func() {
			DeleteYourself(cross)
		})

		if idx != len(adminList)-1 {
			if abort != nil {
				t.Fatal("DeleteYourself didn't remove admins with enough admins")
			}
		} else {
			if abort == nil {
				t.Fatal("DeleteYourself didn't panick while removing last admin")
			}
		}
	}
}

func Test_CountAdmins(t *testing.T) {
	for _, admin := range adminList {
		setDefaultAdmin(std.Address(admin))
	}

	adminNbr := countAdmins()

	if adminNbr != len(adminList) {
		t.Fatal("CountAdmins reported a wrong amount of admins")
	}
}

func Test_IsAdmin(t *testing.T) {
	if len(adminList) < 1 {
		t.Fatal("No admin in adminList")
	}

	owner := std.Address(adminList[len(adminList)-1])

	if !isAdmin(owner) {
		t.Fatal("Last member in adminList not seen as admin")
	}

	owner = std.Address(randomTestWallet)
	testing.SetOriginCaller(owner)

	revive(func() {
		Register(cross)
	})

	if isAdmin(owner) {
		t.Fatal("Random owner seen as admin")
	}
}

func Test_setAdmin(t *testing.T) {
	if len(adminList) < 1 {
		t.Fatal("No admin in adminList")
	}

	owner := std.Address(adminList[len(adminList)-1]) // set owner to admin
	testing.SetOriginCaller(owner)

	abort := revive(func() {
		SetNewAdmin(cross, std.Address(randomTestWallet2))
	})

	if abort == nil {
		t.Fatal("Admin made admin a member that isn't registered")
	}

	owner = std.Address(randomTestWallet2) // set owner to new randomWallet
	testing.SetOriginCaller(owner)

	revive(func() {
		Register(cross)
	})

	owner = std.Address(randomTestWallet) // set owner to normal member
	testing.SetOriginCaller(owner)

	revive(func() {
		Register(cross)
	})

	abort = revive(func() {
		SetNewAdmin(cross, std.Address(randomTestWallet2))
	})

	if abort == nil {
		t.Fatal("Normal user added a new admin")
	}

	owner = std.Address(adminList[len(adminList)-1]) // set owner to admin
	testing.SetOriginCaller(owner)

	abort = revive(func() {
		SetNewAdmin(cross, std.Address(randomTestWallet2))
	})

	if abort != nil {
		t.Fatal("Admin failed to add new admin")
	}

	abort = revive(func() {
		SetNewAdmin(cross, std.Address(randomTestWallet2))
	})

	if abort == nil {
		t.Fatal("Didn't panick while adding the title of admin to an existing admin")
	}

	//remove randomWallet2 from admin for test consistency
	owner = std.Address(randomTestWallet2) // set owner to new randomWallet
	testing.SetOriginCaller(owner)

	revive(func() {
		DeleteYourself(cross)
	})
}

func Test_setDefaultAdmin(t *testing.T) {
	setDefaultAdmin(std.Address(bobAdminWannabe))

	if !isAdmin(std.Address(bobAdminWannabe)) {
		t.Fatal("setDefaultAdmin didn't set a new admin")
	}
}

func Test_banSystem(t *testing.T) {
	t.Run("Ban", func(t *testing.T) {
		owner := std.Address(randomTestWallet)
		testing.SetOriginCaller(owner)

		revive(func() {
			Register(cross)
		})

		owner = std.Address(randomTestWallet2)
		testing.SetOriginCaller(owner)

		revive(func() {
			Register(cross)
		})

		abort := revive(func() {
			BanMember(cross, std.Address(randomTestWallet))
		})

		if abort == nil {
			t.Fatal("Random user banned another user")
		}

		setDefaultAdmin(std.Address(bobAdminWannabe))

		owner = std.Address(adminList[len(adminList)-1]) // set owner to admin
		testing.SetOriginCaller(owner)

		abort = revive(func() {
			BanMember(cross, std.Address(bobAdminWannabe))
		})

		if abort == nil {
			t.Fatal("Admin banned another admin")
		}

		abort = revive(func() {
			BanMember(cross, std.Address(randomTestWallet))
		})

		if abort != nil {
			t.Fatal("Admin can't ban a random user")
		}
	})

	t.Run("unban", func(t *testing.T) {
		owner := std.Address(randomTestWallet2)
		testing.SetOriginCaller(owner)

		abort := revive(func() {
			UnbanMember(cross, std.Address(randomTestWallet))
		})

		if abort == nil {
			t.Fatal("Normal user can unban a banned user")
		}

		owner = std.Address(adminList[len(adminList)-1]) // set user to admin
		testing.SetOriginCaller(owner)

		abort = revive(func() {
			UnbanMember(cross, std.Address(randomTestWallet))
		})

		if abort != nil {
			t.Fatal("Admin can't unban banned user")
		}

		abort = revive(func() {
			UnbanMember(cross, std.Address(randomTestWallet))
		})

		if abort == nil {
			t.Fatal("Admin can unban a user that isn't banned")
		}
	})
}
