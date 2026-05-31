package memberstorage_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards/exts/memberstorage"
)

func TestMemberGroupingAdd(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		name := "Foo"
		g := memberstorage.NewMemberGrouping()

		uassert.False(t, g.Has(name), "expect grouping group not to be found")
		uassert.Equal(t, 0, g.Size(), "expect grouping to be empty")
	})

	t.Run("success", func(t *testing.T) {
		name := "Foo"
		g := memberstorage.NewMemberGrouping()
		mg, err := g.Add(name)

		urequire.NoError(t, err, "expect no error")
		uassert.True(t, g.Has(name), "expect grouping group to be found")
		uassert.Equal(t, 1, g.Size(), "expect grouping to have a single group")

		urequire.True(t, mg != nil, "expected grouping group to be not nil")
		uassert.Equal(t, name, mg.Name(), "expect group to have the right name")
	})

	t.Run("duplicated name", func(t *testing.T) {
		name := "Foo"
		g := memberstorage.NewMemberGrouping()
		_, err := g.Add(name)
		urequire.NoError(t, err, "expect no error")

		_, err = g.Add(name)
		uassert.ErrorContains(t, err, "member group already exists: Foo", "expect duplication error")
	})
}

func TestMemberGroupingGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		name := "Foo"
		g := memberstorage.NewMemberGrouping()
		g.Add(name)

		mg, found := g.Get(name)

		urequire.True(t, found, "expect grouping group to be found")
		urequire.True(t, mg != nil, "expect grouping group to be not nil")
		uassert.Equal(t, name, mg.Name(), "expect group to have the right name")
	})

	t.Run("group not found", func(t *testing.T) {
		g := memberstorage.NewMemberGrouping()

		_, found := g.Get("Foo")

		urequire.False(t, found, "expect grouping group to be not found")
	})
}

func TestMemberGroupingDelete(t *testing.T) {
	name := "Foo"
	g := memberstorage.NewMemberGrouping()
	g.Add(name)

	err := g.Delete(name)

	urequire.NoError(t, err, "expect no error")
	uassert.False(t, g.Has(name), "expect grouping group not to be found")
}

func TestMemberGroupingIterate(t *testing.T) {
	groups := []string{"Tier 1", "Tier 2", "Tier 3"}
	g := memberstorage.NewMemberGrouping()
	for _, name := range groups {
		g.Add(name)
	}

	var i int
	g.IterateByOffset(0, g.Size(), func(mg memberstorage.MemberGroup) bool {
		urequire.True(t, mg != nil, "expect member group not to be nil")
		urequire.Equal(t, groups[i], mg.Name(), "expect group to be iterated in order")

		i++
		return false
	})

	uassert.Equal(t, len(groups), i, "expect all groups to be iterated")
}

func TestIsCanonicalMemberGrouping(t *testing.T) {
	cases := []struct {
		name  string
		setup func() memberstorage.MemberGrouping
		want  bool
	}{
		{
			name:  "canonical",
			setup: func() memberstorage.MemberGrouping { return memberstorage.NewMemberGrouping() },
			want:  true,
		},
		{
			name:  "non canonical",
			setup: func() memberstorage.MemberGrouping { return stubMemberGrouping{} },
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uassert.Equal(t, tc.want, memberstorage.IsCanonicalMemberGrouping(tc.setup()))
		})
	}
}

type stubMemberGrouping struct{}

func (stubMemberGrouping) Size() int                                     { return 0 }
func (stubMemberGrouping) Has(string) bool                               { return false }
func (stubMemberGrouping) Add(string) (memberstorage.MemberGroup, error) { return nil, nil }
func (stubMemberGrouping) Get(string) (memberstorage.MemberGroup, bool)  { return nil, false }
func (stubMemberGrouping) Delete(string) error                           { return nil }

func (stubMemberGrouping) IterateByOffset(offset, count int, fn memberstorage.MemberGroupIterFn) bool {
	return false
}
