package grc721

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

type evilTeller struct{ Teller }

func TestIsCanonicalTeller(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)

	var realmTeller Teller
	func(cur realm) {
		realmTeller = tok.RealmTeller(0, cur)
	}(cross(cur))

	tests := []struct {
		name string
		te   Teller
		want bool
	}{
		{"CallerTeller is canonical", led.CallerTeller(), true},
		{"ReadonlyTeller is canonical", tok.ReadonlyTeller(), true},
		{"RealmTeller is canonical", realmTeller, true},
		{"ImpersonateTeller is canonical", led.ImpersonateTeller(alice), true},
		{"embedding wrapper is rejected", &evilTeller{led.CallerTeller()}, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			uassert.Equal(t, tc.want, IsCanonicalTeller(tc.te))
		})
	}
}

func TestTellerNilGuards(cur realm, t *testing.T) {
	var nilTok *Token
	var nilLed *PrivateLedger

	tests := []struct {
		name string
		msg  string
		fn   func()
	}{
		{"CallerTeller on nil ledger panics", "Ledger cannot be nil", func() {
			nilLed.CallerTeller()
		}},
		{"ReadonlyTeller on nil token panics", "Token cannot be nil", func() {
			nilTok.ReadonlyTeller()
		}},
		{"RealmTeller on nil token panics", "Token cannot be nil", func() {
			nilTok.RealmTeller(0, cur)
		}},
		{"RealmSubTeller on nil token panics", "Token cannot be nil", func() {
			nilTok.RealmSubTeller(0, cur, "slug")
		}},
		{"ImpersonateTeller on nil ledger panics", "Ledger cannot be nil", func() {
			nilLed.ImpersonateTeller(alice)
		}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			uassert.PanicsWithMessage(t, cur, tc.msg, tc.fn)
		})
	}
}

func TestReadonlyTeller(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)
	urequire.NoError(t, led.Mint(alice, "1"))
	ro := tok.ReadonlyTeller()

	tests := []struct {
		name string
		op   string
	}{
		{"Approve rejected with ErrReadonly", "approve"},
		{"TransferFrom rejected with ErrReadonly", "transfer"},
		{"SetApprovalForAll rejected with ErrReadonly", "setall"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testing.SetRealm(testing.NewUserRealm(alice))
			func(cur realm) {
				switch tc.op {
				case "approve":
					err = ro.Approve(0, cur, bob, "1")
				case "transfer":
					err = ro.TransferFrom(0, cur, alice, bob, "1")
				case "setall":
					err = ro.SetApprovalForAll(0, cur, bob, true)
				}
			}(cross(cur))
			uassert.ErrorIs(t, err, ErrReadonly)
		})
	}
}

func TestCallerTellerFlow(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)
	urequire.NoError(t, led.Mint(alice, "1"))
	urequire.NoError(t, led.Mint(alice, "2"))

	teller := led.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, teller.Approve(0, cur, bob, "1"))
	}(cross(cur))

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, teller.SetApprovalForAll(0, cur, bob, true))
	}(cross(cur))
	uassert.True(t, tok.IsApprovedForAll(alice, bob))

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		urequire.NoError(t, teller.TransferFrom(0, cur, alice, carl, "1"))
	}(cross(cur))

	owner, err := tok.OwnerOf("1")
	urequire.NoError(t, err)
	uassert.Equal(t, carl, owner)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorIs(t, teller.TransferFrom(0, cur, carl, bob, "1"), ErrCallerIsNotOwnerOrApproved)
	}(cross(cur))
}

func TestCallerTellerSpoofedRealm(cur realm, t *testing.T) {
	_, led := newTestToken("Foo", "FOO", 0, cur)
	urequire.NoError(t, led.Mint(alice, "1"))
	teller := led.CallerTeller()

	// Stale outer cur while a fresh frame is current → ErrSpoofedRealm.
	tests := []struct {
		name string
		op   string
	}{
		{"Approve rejects spoofed realm", "approve"},
		{"TransferFrom rejects spoofed realm", "transfer"},
		{"SetApprovalForAll rejects spoofed realm", "setall"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var err error
			stale := cur
			func(cur realm) {
				switch tc.op {
				case "approve":
					err = teller.Approve(0, stale, bob, "1")
				case "transfer":
					err = teller.TransferFrom(0, stale, alice, bob, "1")
				case "setall":
					err = teller.SetApprovalForAll(0, stale, bob, true)
				}
			}(cross(cur))
			uassert.ErrorIs(t, err, ErrSpoofedRealm)
		})
	}
}

func TestRealmTellerFlow(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)

	var teller Teller
	var realmAddr address
	func(cur realm) {
		teller = tok.RealmTeller(0, cur)
		realmAddr = cur.Address()
	}(cross(cur))

	urequire.NoError(t, led.Mint(realmAddr, "1"))

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, teller.TransferFrom(0, cur, realmAddr, bob, "1"))
	}(cross(cur))

	owner, err := tok.OwnerOf("1")
	urequire.NoError(t, err)
	uassert.Equal(t, bob, owner)
}

func TestRealmSubTellerFlow(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)

	tests := []struct {
		name           string
		slug           string
		tid            TokenID
		wantsRealmAddr bool
	}{
		{"empty slug derives the realm address itself", "", "10", true},
		{"non-empty slug derives a sub-account", "vault", "11", false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var teller Teller
			var realmAddr, account address
			func(cur realm) {
				teller = tok.RealmSubTeller(0, cur, tc.slug)
				realmAddr = cur.Address()
				account = accountSlugAddr(cur.Address(), tc.slug)
			}(cross(cur))

			if tc.wantsRealmAddr {
				uassert.Equal(t, realmAddr, account)
			} else {
				uassert.True(t, realmAddr != account, "sub-account must differ from realm address")
			}

			urequire.NoError(t, led.Mint(account, tc.tid))
			testing.SetRealm(testing.NewUserRealm(alice))
			func(cur realm) {
				urequire.NoError(t, teller.TransferFrom(0, cur, account, bob, tc.tid))
			}(cross(cur))

			owner, err := tok.OwnerOf(tc.tid)
			urequire.NoError(t, err)
			uassert.Equal(t, bob, owner)
		})
	}
}

func TestRealmTellerSpoofedRealm(cur realm, t *testing.T) {
	tok, _ := newTestToken("Foo", "FOO", 0, cur)

	tests := []struct {
		name string
		fn   func()
	}{
		{"RealmTeller rejects spoofed realm", func() {
			stale := cur
			func(cur realm) {
				tok.RealmTeller(0, stale)
			}(cross(cur))
		}},
		{"RealmSubTeller rejects spoofed realm", func() {
			stale := cur
			func(cur realm) {
				tok.RealmSubTeller(0, stale, "vault")
			}(cross(cur))
		}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			urequire.AbortsContains(t, cur, "rlm does not match the current crossing frame", tc.fn)
		})
	}
}

func TestImpersonateTeller(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)
	urequire.NoError(t, led.Mint(alice, "1"))

	teller := led.ImpersonateTeller(alice)
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		urequire.NoError(t, teller.TransferFrom(0, cur, alice, bob, "1"))
	}(cross(cur))

	owner, err := tok.OwnerOf("1")
	urequire.NoError(t, err)
	uassert.Equal(t, bob, owner)
}

func TestImpersonateTellerRejectsInvalidAddress(cur realm, t *testing.T) {
	_, led := newTestToken("Foo", "FOO", 0, cur)
	uassert.PanicsWithMessage(t, cur, "invalid address", func() {
		led.ImpersonateTeller(zeroAddress)
	})
}
