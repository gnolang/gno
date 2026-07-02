package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

func TestMemberGroupNew(cur realm, t *testing.T) {
	urequire.PanicsWithMessage(t, cur, "member storage is required", func() {
		commondao.NewMemberGroup("", nil)
	})

	storage := commondao.NewMemberStorage()
	urequire.PanicsWithMessage(t, cur, "member group name is required", func() {
		commondao.NewMemberGroup("", storage)
	})

	urequire.NotPanics(t, cur, func() {
		name := "Tier 1"
		g := commondao.NewMemberGroup(name, storage)
		uassert.Equal(t, name, g.Name(), "expect group name to match")
		uassert.NotNil(t, g.Members(), "expect members to be not nil")
		uassert.Nil(t, g.GetMeta(), "expect default group meta to be nil")
	})
}

func TestMemberGroupSetMeta(cur realm, t *testing.T) {
	cases := []struct {
		name     string
		wantMeta any
		setup    func() commondao.MemberGroup
		panicMsg string
	}{
		{
			name:     "ok",
			wantMeta: 42,
			setup: func() commondao.MemberGroup {
				return commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
			},
		},
		{
			name:     "overwrite",
			wantMeta: 42,
			setup: func() commondao.MemberGroup {
				g := commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
				g.SetMeta(123)
				return g
			},
		},
		{
			name:     "sealed group",
			wantMeta: 1,
			setup: func() commondao.MemberGroup {
				g := commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
				return g.Seal().(commondao.MemberGroup)
			},
			panicMsg: "type is sealed so it can't be modified",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			g := tc.setup()

			// Act
			fn := func() {
				g.SetMeta(tc.wantMeta)
			}

			// Assert
			if tc.panicMsg != "" {
				urequire.PanicsWithMessage(t, cross(cur), tc.panicMsg, fn)
			} else {
				urequire.NotPanics(t, cross(cur), fn)
				uassert.True(t, g.GetMeta() == tc.wantMeta, "expect metadata to match")
			}
		})
	}
}

func TestMemberGroupGetMeta(cur realm, t *testing.T) {
	cases := []struct {
		name     string
		wantMeta any
		setup    func() commondao.MemberGroup
		panicMsg string
	}{
		{
			name:     "ok",
			wantMeta: 42,
			setup: func() commondao.MemberGroup {
				g := commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
				g.SetMeta(42)
				return g
			},
		},
		{
			name: "sealed group",
			setup: func() commondao.MemberGroup {
				g := commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
				g.SetMeta(42)
				return g.Seal().(commondao.MemberGroup)
			},
			panicMsg: "failed to get metadata, group is sealed and metadata is not sealable",
		},
		{
			name:     "sealed group with sealable value",
			wantMeta: sealableInt(-42), // Negative when sealed
			setup: func() commondao.MemberGroup {
				g := commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
				g.SetMeta(sealableInt(42))
				return g.Seal().(commondao.MemberGroup)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var got any
			g := tc.setup()

			// Act
			fn := func() {
				got = g.GetMeta()
			}

			// Assert
			if tc.panicMsg != "" {
				urequire.PanicsWithMessage(t, cross(cur), tc.panicMsg, fn)
			} else {
				urequire.NotPanics(t, cross(cur), fn)
				uassert.True(t, got == tc.wantMeta, "expect metadata to match")
			}
		})
	}
}

type sealableInt int

func (i sealableInt) Seal() commondao.Sealable { return -i } // Negative when sealed
func (sealableInt) IsSealed() bool             { return true }
