package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

func TestMemberGroupingAdd(t *testing.T) {
	t.Run("defauls", func(t *testing.T) {
		name := "Foo"
		g := commondao.NewMemberGrouping()

		uassert.False(t, g.Has(name), "expect grouping group not to be found")
		uassert.Equal(t, 0, g.Size(), "expect grouping to be empty")
	})

	t.Run("success", func(t *testing.T) {
		name := "Foo"
		g := commondao.NewMemberGrouping()
		mg, err := g.Add(name)

		urequire.NoError(t, err, "expect no error")
		uassert.True(t, g.Has(name), "expect grouping group to be found")
		uassert.Equal(t, 1, g.Size(), "expect grouping to have a single group")

		urequire.True(t, mg != nil, "expected grouping group to be not nil")
		uassert.Equal(t, name, mg.Name(), "expect group to have the right name")
	})

	t.Run("duplicated name", func(t *testing.T) {
		name := "Foo"
		g := commondao.NewMemberGrouping()
		_, err := g.Add(name)
		urequire.NoError(t, err, "expect no error")

		_, err = g.Add(name)
		uassert.ErrorContains(t, err, "member group already exists: Foo", "expect duplication error")
	})
}

func TestMemberGroupingGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		name := "Foo"
		g := commondao.NewMemberGrouping()
		g.Add(name)

		mg, found := g.Get(name)

		urequire.True(t, found, "expect grouping group to be found")
		urequire.True(t, mg != nil, "expect grouping group to be not nil")
		uassert.Equal(t, name, mg.Name(), "expect group to have the right name")
	})

	t.Run("group not found", func(t *testing.T) {
		g := commondao.NewMemberGrouping()

		_, found := g.Get("Foo")

		urequire.False(t, found, "expect grouping group to be not found")
	})
}

func TestMemberGroupingDelete(t *testing.T) {
	name := "Foo"
	g := commondao.NewMemberGrouping()
	g.Add(name)

	err := g.Delete(name)

	urequire.NoError(t, err, "expect no error")
	uassert.False(t, g.Has(name), "expect grouping group not to be found")
}

func TestMemberGroupingIterate(t *testing.T) {
	groups := []string{"Tier 1", "Tier 2", "Tier 3"}
	g := commondao.NewMemberGrouping()
	for _, name := range groups {
		g.Add(name)
	}

	var i int
	g.IterateByOffset(0, g.Size(), func(mg commondao.MemberGroup) bool {
		urequire.True(t, mg != nil, "expect member group not to be nil")
		urequire.Equal(t, groups[i], mg.Name(), "expect group to be iterated in order")

		i++
		return false
	})

	uassert.Equal(t, len(groups), i, "expect all groups to be iterated")
}

func TestMemberGroupingUseStorageFactory(t *testing.T) {
	var (
		groupName     string
		defaultMember = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	)

	fn := func(group string) commondao.MemberStorage {
		groupName = group

		// Create storage and add a default member
		s := commondao.NewMemberStorage()
		s.Add(defaultMember)
		return s
	}

	grouping := commondao.NewMemberGrouping(commondao.UseStorageFactory(fn))
	group, err := grouping.Add("foo")

	urequire.NoError(t, err, "expect no group creation error")
	uassert.True(t, group.Members().Has(defaultMember), "expect storage to have a default member")
	uassert.Equal(t, "foo", groupName, "expect group name to match")
}

func TestMemberGroupingWithGroups(t *testing.T) {
	defaultGroups := []string{"foo", "bar", "baz"}
	grouping := commondao.NewMemberGrouping(commondao.WithGroups(defaultGroups...))

	urequire.Equal(t, len(defaultGroups), grouping.Size(), "expect groups to be created")
	for _, group := range defaultGroups {
		uassert.True(t, grouping.Has(group), "expect group to be found: "+group)
	}
}
