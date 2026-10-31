package gnolang

import "testing"

func TestObjectID_UnmarshalAmino_InvalidPkgIDLength(t *testing.T) {
	var oid ObjectID
	if oid.UnmarshalAmino("abc:100") == nil {
		t.Error("expected error for short PkgID")
	}
}

func TestObjectID_UnmarshalAmino_NewTimeZero(t *testing.T) {
	var oid ObjectID
	if oid.UnmarshalAmino("abcdef0123456789abcdef0123456789abcdef01:0") == nil {
		t.Error("expected error for NewTime=0 with non-zero PkgID")
	}
}

func TestObjectID_UnmarshalAmino_NegativeNewTime(t *testing.T) {
	var oid ObjectID
	if oid.UnmarshalAmino("abcdef0123456789abcdef0123456789abcdef01:-1") == nil {
		t.Error("expected error for negative NewTime")
	}
}
