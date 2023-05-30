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

func TestCategoryEvaluationCriteria(t *testing.T) {
	member := testutils.TestAddress("member")
	category := "document"
	criteria := []string{"clarity", "usage"}
	category2 := "bounty"
	criteria2 := []string{"complexity"}
	expectedGategory := NewCategory(category, criteria)
	expectedGategory2 := NewCategory(category2, criteria2)

	c := NewCommity()
	c.DesignateMembers([]std.Address{member})

	t.Run("Add First Commity Category and Evaluation Criteria", func(t *testing.T) {
		std.TestSetOrigCaller(member)
		c.AddCategory(category, criteria)
		value, exists := c.categories.Get(category)
		if !exists {
			t.Errorf("Add first category %s failed", category)
		}
		gotCategory := value.(*Category)
		if gotCategory.name != expectedGategory.name {
			t.Errorf("First Commity category got %s expected %s", gotCategory.name, expectedGategory.name)
		}
	})

	t.Run("Add Second Commity Category and Evaluation Criteria", func(t *testing.T) {
		std.TestSetOrigCaller(member)
		c.AddCategory(category2, criteria2)
		value2, exists2 := c.categories.Get(category2)
		if !exists2 {
			t.Errorf("Add second category %s failed", category2)
		}
		gotCategory2 := value2.(*Category)
		if gotCategory2.name != expectedGategory2.name {
			t.Errorf("Second Commity category got %s expected %s", gotCategory2.name, expectedGategory2.name)
		}
	})

	t.Run("Approve First Commity Category", func(t *testing.T) {
		std.TestSetOrigCaller(member)
		approved := c.ApproveCategory(category, VOTE_YES)
		if !approved {
			value, exists := c.categories.Get(category)
			gotCategory := value.(*Category)
			t.Errorf("Approved First Commity category got %s expected %s", gotCategory.status, "Approved")
		}
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
