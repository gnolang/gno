package bptree

import (
	"errors"
	"testing"
)

// TestMutableTree_IterateResolverErrorPropagates verifies that when
// resolveValue fails during Iterate, the error is returned rather than
// silently dropped. See Finding #8.
func TestMutableTree_IterateResolverErrorPropagates(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 5; i++ {
		if _, err := tree.Set([]byte{byte(i)}, []byte{byte(i + 100)}); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}

	// Corrupt the value store by removing one value. The tree still
	// references its ValueKey via the leaf, so Iterate will hit a
	// missing-value error when it tries to resolve.
	var dropVK string
	for vk := range tree.memValues {
		dropVK = vk
		break
	}
	delete(tree.memValues, dropVK)

	_, err := tree.Iterate(func(k, v []byte) bool { return false })
	if err == nil {
		t.Fatal("Iterate err = nil; want error from broken resolver")
	}
}

// TestImmutableTree_IterateResolverErrorPropagates verifies the same
// for ImmutableTree.Iterate.
func TestImmutableTree_IterateResolverErrorPropagates(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 5; i++ {
		if _, err := tree.Set([]byte{byte(i)}, []byte{byte(i + 100)}); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}

	imm := tree.Snapshot(0)
	failErr := errors.New("synthetic resolver failure")
	imm.SetValueResolver(func(vk []byte) ([]byte, error) {
		return nil, failErr
	})

	_, err := imm.Iterate(func(k, v []byte) bool { return false })
	if !errors.Is(err, failErr) {
		t.Fatalf("Iterate err = %v; want %v", err, failErr)
	}
}

// TestMutableTree_IterateEmptyTreeReturnsNilError is a sanity test that
// an empty tree's Iterate returns (false, nil) — the error path must
// not regress the happy path.
func TestMutableTree_IterateEmptyTreeReturnsNilError(t *testing.T) {
	tree := NewMutableTreeMem()
	stopped, err := tree.Iterate(func(k, v []byte) bool { return false })
	if err != nil {
		t.Fatalf("empty Iterate err = %v; want nil", err)
	}
	if stopped {
		t.Fatalf("empty Iterate stopped = true; want false")
	}
}

// TestMutableTree_IterateSuccessPath ensures the happy path still
// returns (false, nil) when all resolutions succeed.
func TestMutableTree_IterateSuccessPath(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 10; i++ {
		if _, err := tree.Set([]byte{byte(i)}, []byte{byte(i + 100)}); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}
	count := 0
	stopped, err := tree.Iterate(func(k, v []byte) bool {
		count++
		return false
	})
	if err != nil {
		t.Fatalf("Iterate err = %v; want nil", err)
	}
	if stopped {
		t.Fatalf("Iterate stopped unexpectedly")
	}
	if count != 10 {
		t.Fatalf("Iterate count = %d; want 10", count)
	}
}
