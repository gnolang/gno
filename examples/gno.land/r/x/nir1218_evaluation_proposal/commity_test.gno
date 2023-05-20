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
	category := "document"
	criteria := []string{"clarity", "usage"}
	category2 := "bounty"
	criteria2 := []string{"complexity"}
	// expectedCategories := []string{category}
	// expectedCategories2 := []string{category2, category}

	c := NewCommity()

	t.Run("Add First Commity Category and Evaluation Criteria", func(t *testing.T) {
		c.AddCategory(category, criteria)
		// if !isEqualStringSlice(gotCategories, expectedCategories) {
		// 	t.Errorf("Commity categoriess got %v expcted %v", gotCategories, expectedCategories)
		// }
	})

	t.Run("Add Second Commity Category and Evaluation Criteria", func(t *testing.T) {
		c.AddCategory(category2, criteria2)
		// if !isEqualStringSlice(gotCategories2, expectedCategories2) {
		// 	t.Errorf("Commity categoriess got %v expcted %v", gotCategories2, expectedCategories2)
		// }
	})

	t.Run("Approve Category", func(t *testing.T) {
		c.ApproveCategory(category2, "YES")
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
