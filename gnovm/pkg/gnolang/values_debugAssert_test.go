//go:build debugAssert

package gnolang

import (
	"testing"
)

func TestFillValueTVHashChain(t *testing.T) {
	oid := ObjectID{
		PkgID:   PkgID{Hashlet: NewHashlet(seed20('A'))},
		NewTime: 1,
	}

	av := &ArrayValue{}
	av.SetObjectID(oid)
	correctHash := ValueHash{NewHashlet(seed20('X'))}
	av.SetHash(correctHash)

	store := &mockStore{obj: av}

	// TypedValue holds a RefValue with a WRONG hash.
	wrongHash := ValueHash{NewHashlet(seed20('W'))}
	tv := TypedValue{
		T: nil,
		V: RefValue{
			ObjectID: oid,
			Hash:     wrongHash,
		},
	}

	caught := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				caught = true
			}
		}()
		fillValueTV(store, &tv)
	}()

	if !caught {
		t.Fatal("expected panic: hash chain broken, but fillValueTV returned normally")
	}
}

func TestFillValueTVHashChainMatch(t *testing.T) {
	oid := ObjectID{
		PkgID:   PkgID{Hashlet: NewHashlet(seed20('B'))},
		NewTime: 2,
	}

	av := &ArrayValue{}
	av.SetObjectID(oid)
	correctHash := ValueHash{NewHashlet(seed20('X'))}
	av.SetHash(correctHash)

	store := &mockStore{obj: av}

	// TypedValue holds a RefValue with the CORRECT hash.
	tv := TypedValue{
		T: nil,
		V: RefValue{
			ObjectID: oid,
			Hash:     correctHash,
		},
	}

	fillValueTV(store, &tv) // should not panic
}

func TestFillValueTVZeroHash(t *testing.T) {
	oid := ObjectID{
		PkgID:   PkgID{Hashlet: NewHashlet(seed20('C'))},
		NewTime: 3,
	}

	av := &ArrayValue{}
	av.SetObjectID(oid)
	av.SetHash(ValueHash{NewHashlet(seed20('Y'))})

	store := &mockStore{obj: av}

	// Zero hash skips the check (escaped objects resolve via IAVL).
	tv := TypedValue{
		T: nil,
		V: RefValue{
			ObjectID: oid,
			Hash:     ValueHash{},
		},
	}

	fillValueTV(store, &tv) // should not panic
}
