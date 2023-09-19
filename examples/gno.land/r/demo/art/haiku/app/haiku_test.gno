package haiku

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAmulet(t *testing.T) {
	// 	haiku := `yes web three is here
	// 	now you can own your data
	// the future is bright`
	haiku := `a zoo a zoo a
zoo a zoo a zoo a zoo 
a zoo a zoo a`
	haiku = `a zoo a zoo a\nzoo a zoo a zoo a zoo \na zoo a zoo a\n`
	if err := isHaiku(haiku); err != nil {
		t.Errorf("not haiku: %s", err.Error())
	}
	if calculateRarity("88aa888aaaa8a8a8a8a88a8a") != "common" {
		t.Errorf("rarity calculations errored")
	}
	addr1 := testutils.TestAddress("addr1")
	addr2 := testutils.TestAddress("addr2")
	std.TestSetOrigCaller(addr1)
	id := Mint(haiku)
	if strings.Contains(id, "error:") {
		t.Errorf("could not mint: %s", id)
	}
	if err := Mint("not a haiku"); !strings.Contains(err, "error:") {
		t.Errorf("not a haiku")
	}
	if err := Mint(haiku); !strings.Contains(err, "error:") {
		t.Errorf("token already exists")
	}
	if err := Transfer(addr2, id); strings.Contains(err, "error:") {
		t.Errorf("could not transfer: %s", err)
	}
	if err := Transfer(addr2, id); !strings.Contains(err, "error:") {
		t.Errorf("should not be able to transfer: %s", err)
	}

	// println(Transfer(addr2, "ae60a376a9be58296c4e31db658f3ea031115553169964f9c5b54f23269704ae"))
	// println("\n\n~~~~~~~~~~~\n\n")
	// println(Render(""))
	// println("\n\n~~~~~~~~~~~\n\n")
	// println(Render("ae60a376a9be58296c4e31db658f3ea031115553169964f9c5b54f23269704ae"))
	// println("\n\n~~~~~~~~~~~\n\n")
	// println(Render("createdby:" + addr1.String()))
	// println("\n\n~~~~~~~~~~~\n\n")
	// println(Render("collectedby:" + addr2.String()))
}
