package evaluation

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestCommityMembers(t *testing.T) {
	members := []std.Address{testutils.TestAddress("member1"), testutils.TestAddress("member2"), testutils.TestAddress("member3")}
	c := NewCommity()

	t.Run("Designate Commity Members", func(t *testing.T) {
		c.DesignateMembers(members)
		if !isEqualAddressSlice(c.members, members) {
			t.Errorf("Designated Commity members got %v expcted %v", members, c.members)
		}
	})

	t.Run("Dismiss Commity Members", func(t *testing.T) {
		c.DismissMembers(members)
	})
}

func TestCommityCategories(t *testing.T) {
	categories := []string{"bounty", "chore", "defct", "document", "feat", "tutorial"}
	c := NewCommity()

	t.Run("Add Commity Categories", func(t *testing.T) {
		updatedCategories := c.AddCategories(categories)
		if !isEqualStringSlice(updatedCategories, categories) {
			t.Errorf("Commity categoriess got %v expcted %v", updatedCategories, categories)
		}
	})

	t.Run("Add Commity Categories", func(t *testing.T) {
		c.RemoveCategories(categories)
	})
}

func isEqualStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isEqualAddressSlice(a, b []std.Address) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
