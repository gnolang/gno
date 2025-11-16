package gnolang

import "testing"

func TestObjectID_UnmarshalAmino_InvalidPkgIDLength(t *testing.T) {
	var oid ObjectID

	// Too short
	err := oid.UnmarshalAmino("abc:100")
	if err == nil {
		t.Error("expected error for short PkgID")
	}

	// Too long
	err = oid.UnmarshalAmino("abcdef0123456789abcdef0123456789abcdef01234:100")
	if err == nil {
		t.Error("expected error for long PkgID")
	}

	// Valid length should work
	err = oid.UnmarshalAmino("abcdef0123456789abcdef0123456789abcdef01:100")
	if err != nil {
		t.Errorf("unexpected error for valid PkgID: %v", err)
	}
}

func TestObjectID_UnmarshalAmino_NewTimeZeroWithNonZeroPkgID(t *testing.T) {
	var oid ObjectID

	// NewTime zero with non-zero PkgID should fail
	err := oid.UnmarshalAmino("abcdef0123456789abcdef0123456789abcdef01:0")
	if err == nil {
		t.Error("expected error for NewTime=0 with non-zero PkgID")
	}

	// NewTime zero with zero PkgID should work
	err = oid.UnmarshalAmino("0000000000000000000000000000000000000000:0")
	if err != nil {
		t.Errorf("unexpected error for NewTime=0 with zero PkgID: %v", err)
	}

	// NewTime > 0 with non-zero PkgID should work
	err = oid.UnmarshalAmino("abcdef0123456789abcdef0123456789abcdef01:1")
	if err != nil {
		t.Errorf("unexpected error for NewTime>0 with non-zero PkgID: %v", err)
	}
}
