//go:build debugAssert

package gnolang

import (
	"testing"
)

// mockStore implements Store just enough for the RefValue DeepFill hash chain test.
type mockStore struct {
	Store // nil — panics if any other method is called
	obj   Object
}

func (m *mockStore) GetObject(oid ObjectID) Object { return m.obj }

func newOID(seed byte, n uint64) ObjectID {
	return ObjectID{
		PkgID:   PkgID{Hashlet: NewHashlet(seed20(seed))},
		NewTime: n,
	}
}

func seed20(seed byte) []byte {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = seed
	}
	return bz
}

func TestRefValueDeepFillHashChain(t *testing.T) {
	oid := newOID('a', 1)

	// Object with a known hash.
	av := &ArrayValue{}
	av.SetObjectID(oid)
	correctHash := ValueHash{NewHashlet(seed20('X'))}
	av.SetHash(correctHash)

	store := &mockStore{obj: av}

	// RefValue with a DIFFERENT hash than the child's.
	wrongHash := ValueHash{NewHashlet(seed20('W'))}
	rv := RefValue{
		ObjectID: oid,
		Hash:     wrongHash,
	}

	caught := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				caught = true
			}
		}()
		rv.DeepFill(store)
	}()

	if !caught {
		t.Fatal("expected panic: hash chain broken, but DeepFill returned normally")
	}
}

func TestRefValueDeepFillHashChainMatch(t *testing.T) {
	oid := newOID('b', 2)

	av := &ArrayValue{}
	av.SetObjectID(oid)
	correctHash := ValueHash{NewHashlet(seed20('X'))}
	av.SetHash(correctHash)

	store := &mockStore{obj: av}

	// RefValue with the CORRECT hash — should not panic.
	rv := RefValue{
		ObjectID: oid,
		Hash:     correctHash,
	}

	_ = rv.DeepFill(store)
}

func TestRefValueDeepFillZeroHash(t *testing.T) {
	oid := newOID('c', 3)

	av := &ArrayValue{}
	av.SetObjectID(oid)
	av.SetHash(ValueHash{NewHashlet(seed20('Y'))})

	store := &mockStore{obj: av}

	// Zero hash skips the check (escaped objects resolve via IAVL).
	rv := RefValue{
		ObjectID: oid,
		Hash:     ValueHash{},
	}

	_ = rv.DeepFill(store) // should not panic
}
