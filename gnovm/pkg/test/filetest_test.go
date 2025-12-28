package test

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestProvisionFiletestAllocator(t *testing.T) {
	t.Run("no allocator needed", func(t *testing.T) {
		m := &gno.Machine{}
		provisionFiletestAllocator(m, false)
		if m.Alloc != nil {
			t.Fatalf("expected allocator to remain nil")
		}
	})

	t.Run("allocates when needed", func(t *testing.T) {
		m := &gno.Machine{}
		provisionFiletestAllocator(m, true)
		if m.Alloc == nil {
			t.Fatalf("expected allocator to be provisioned")
		}
	})

	t.Run("preserves existing allocator", func(t *testing.T) {
		m := &gno.Machine{Alloc: gno.NewAllocator(128)}
		ptr := m.Alloc
		provisionFiletestAllocator(m, true)
		if m.Alloc != ptr {
			t.Fatalf("expected allocator to remain unchanged")
		}
	})
}
