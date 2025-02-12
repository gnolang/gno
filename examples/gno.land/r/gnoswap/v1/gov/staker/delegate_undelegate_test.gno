package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
)

func TestDelegateInternal(t *testing.T) {
	t.Skip("Must running separately. because this test depends on TestUndelegate's state")
	std.TestSetOrigCaller(testAddr1)
	resetState()

	addr1 := testAddr1

	t.Run("success - first delegation", func(t *testing.T) {
		delegate(addr1, 100)

		if totalDelegated != 100 {
			t.Errorf("expected totalDelegated 100, got %d", totalDelegated)
		}

		value, exists := delegatorAmount.Get(addr1.String())
		if !exists || value.(uint64) != 100 {
			t.Error("delegator amount not updated correctly")
		}

		innerTree, exists := delegatedFromTo.Get(addr1.String())
		if !exists {
			t.Error("delegatedFromTo not updated")
		}

		inner := innerTree.(*avl.Tree)
		delegatedAmount, exists := inner.Get(addr1.String())
		if !exists {
			t.Error("delegatedFromTo amount incorrect")
		}

		if delegatedAmount.(uint64) != 100 {
			t.Error("delegatedFromTo amount incorrect")
		}
	})

	t.Run("success - additional delegation", func(t *testing.T) {
		resetState()
		delegate(addr1, 100)
		delegate(addr1, 50)

		if totalDelegated != 150 {
			t.Errorf("expected totalDelegated 150, got %d", totalDelegated)
		}
	})
}

func TestUndelegateInternal(t *testing.T) {
	t.Skip("Must running separately. because this test depends on TestUndelegate's state")
	addr1 := testAddr1
	std.TestSetOrigCaller(addr1)
	resetState()

	t.Run("fail - no delegation", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for no delegation")
			}
		}()
		undelegate(addr1, 100)
	})

	t.Run("fail - insufficient amount", func(t *testing.T) {
		delegate(addr1, 50)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for insufficient amount")
			}
		}()
		undelegate(addr1, 100)
	})

	t.Run("success - partial undelegate", func(t *testing.T) {
		resetState()
		delegate(addr1, 100)
		undelegate(addr1, 30)

		if totalDelegated != 70 {
			t.Errorf("expected totalDelegated 70, got %d", totalDelegated)
		}
	})

	t.Run("success - full undelegate", func(t *testing.T) {
		resetState()
		delegate(addr1, 100)
		undelegate(addr1, 100)

		if totalDelegated != 0 {
			t.Errorf("expected totalDelegated 0, got %d", totalDelegated)
		}
	})
}

func resetState() {
	totalDelegated = 0
	delegatorAmount = avl.NewTree()
	delegatedFromTo = avl.NewTree()
	delegatedTo = avl.NewTree()
	delegationHistory = avl.NewTree()
	delegationSnapShotHistory = avl.NewTree()
}
