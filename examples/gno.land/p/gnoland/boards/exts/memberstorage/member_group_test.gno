package memberstorage_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards/exts/memberstorage"
)

func TestMemberGroupNew(cur realm, t *testing.T) {
	urequire.PanicsWithMessage(t, cur, "member storage is required", func() {
		memberstorage.NewMemberGroup("", nil)
	})

	storage := memberstorage.New()
	urequire.PanicsWithMessage(t, cur, "member group name is required", func() {
		memberstorage.NewMemberGroup("", storage)
	})

	urequire.NotPanics(t, cur, func() {
		name := "Tier 1"
		g := memberstorage.NewMemberGroup(name, storage)
		uassert.Equal(t, name, g.Name(), "expect group name to match")
		uassert.NotNil(t, g.Members(), "expect members to be not nil")
		uassert.Nil(t, g.GetMeta(), "expect default group meta to be nil")
	})
}

func TestMemberGroupSetMeta(t *testing.T) {
	cases := []struct {
		name     string
		wantMeta any
		setup    func() memberstorage.MemberGroup
	}{
		{
			name:     "ok",
			wantMeta: 42,
			setup: func() memberstorage.MemberGroup {
				return memberstorage.NewMemberGroup("Test", memberstorage.New())
			},
		},
		{
			name:     "overwrite",
			wantMeta: 42,
			setup: func() memberstorage.MemberGroup {
				g := memberstorage.NewMemberGroup("Test", memberstorage.New())
				g.SetMeta(123)
				return g
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(cur realm, t *testing.T) {
			// Arrange
			g := tc.setup()

			// Act
			g.SetMeta(tc.wantMeta)

			// Assert
			uassert.True(t, g.GetMeta() == tc.wantMeta, "expect metadata to match")
		})
	}
}

func TestMemberGroupGetMeta(t *testing.T) {
	cases := []struct {
		name     string
		wantMeta any
		setup    func() memberstorage.MemberGroup
		panicMsg string
	}{
		{
			name:     "ok",
			wantMeta: 42,
			setup: func() memberstorage.MemberGroup {
				g := memberstorage.NewMemberGroup("Test", memberstorage.New())
				g.SetMeta(42)
				return g
			},
		},
		{
			name:     "default",
			wantMeta: nil,
			setup: func() memberstorage.MemberGroup {
				return memberstorage.NewMemberGroup("Test", memberstorage.New())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(cur realm, t *testing.T) {
			// Arrange
			var got any
			g := tc.setup()

			// Act
			got = g.GetMeta()

			// Assert
			uassert.True(t, got == tc.wantMeta, "expect metadata to match")
		})
	}
}

func TestIsCanonical(t *testing.T) {
	cases := []struct {
		name  string
		setup func() memberstorage.MemberGroup
		want  bool
	}{
		{
			name: "canonical",
			setup: func() memberstorage.MemberGroup {
				return memberstorage.NewMemberGroup("Test", memberstorage.New())
			},
			want: true,
		},
		{
			name:  "non canonical",
			setup: func() memberstorage.MemberGroup { return stubMemberGroup{} },
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(cur realm, t *testing.T) {
			uassert.Equal(t, tc.want, memberstorage.IsCanonical(tc.setup()))
		})
	}
}

type stubMemberGroup struct{}

func (stubMemberGroup) Name() string                         { return "" }
func (stubMemberGroup) Members() memberstorage.MemberStorage { return nil }
func (stubMemberGroup) SetMeta(any)                          {}
func (stubMemberGroup) GetMeta() any                         { return nil }
