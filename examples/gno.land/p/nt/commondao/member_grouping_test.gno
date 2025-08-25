package commondao

import (
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestMemberGroupingAdd(t *testing.T) {
	t.Run("defauls", func(t *testing.T) {
		name := "Foo"
		g := NewMemberGrouping()

		uassert.False(t, g.Has(name), "expect grouping group not to be found")
		uassert.Equal(t, 0, g.Size(), "expect grouping to be empty")
	})

	t.Run("success", func(t *testing.T) {
		name := "Foo"
		g := NewMemberGrouping()
		mg, err := g.Add(name)

		urequire.NoError(t, err, "expect no error")
		uassert.True(t, g.Has(name), "expect grouping group to be found")
		uassert.Equal(t, 1, g.Size(), "expect grouping to have a single group")

		urequire.True(t, mg != nil, "expected grouping group to be not nil")
		uassert.Equal(t, name, mg.Name(), "expect group to have the right name")
	})

	t.Run("duplicated name", func(t *testing.T) {
		name := "Foo"
		g := NewMemberGrouping()
		_, err := g.Add(name)
		urequire.NoError(t, err, "expect no error")

		_, err = g.Add(name)
		uassert.ErrorContains(t, err, "member group already exists: Foo", "expect duplication error")
	})
}

func TestMemberGroupingGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		name := "Foo"
		g := NewMemberGrouping()
		g.Add(name)

		mg, found := g.Get(name)

		urequire.True(t, found, "expect grouping group to be found")
		urequire.True(t, mg != nil, "expect grouping group to be not nil")
		uassert.Equal(t, name, mg.Name(), "expect group to have the right name")
	})

	t.Run("group not found", func(t *testing.T) {
		g := NewMemberGrouping()

		_, found := g.Get("Foo")

		urequire.False(t, found, "expect grouping group to be not found")
	})
}

func TestMemberGroupingDelete(t *testing.T) {
	name := "Foo"
	g := NewMemberGrouping()
	g.Add(name)

	err := g.Delete(name)

	uassert.NoError(t, err, "expect no error")
	urequire.False(t, g.Has(name), "expect grouping group not to be found")
}

func TestMemberGroupingIterate(t *testing.T) {
	groups := []string{"Tier 1", "Tier 2", "Tier 3"}
	g := NewMemberGrouping()
	for _, name := range groups {
		g.Add(name)
	}

	var i int
	g.IterateByOffset(0, g.Size(), func(mg MemberGroup) bool {
		urequire.True(t, mg != nil, "expect member group not to be nil")
		urequire.Equal(t, groups[i], mg.Name(), "expect group to be iterated in order")

		i++
		return false
	})

	uassert.Equal(t, len(groups), i, "expect all groups to be iterated")
}
