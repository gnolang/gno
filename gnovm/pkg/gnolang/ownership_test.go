package gnolang

import (
	"math"
	"strings"
	"testing"
)

func TestObjectID_UnmarshalAmino_Invalid(t *testing.T) {
	const pkgID = "abcdef0123456789abcdef0123456789abcdef01"
	for _, tc := range []struct{ oids, wantErr string }{
		{"abc:100", "invalid hash length"},
		{pkgID + ":-1", "strconv.ParseUint"},
		{pkgID + ":+1", "strconv.ParseUint"},
		{pkgID + ":0x1", "strconv.ParseUint"},
	} {
		var oid ObjectID
		err := oid.UnmarshalAmino(tc.oids)
		if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%q: got error %v, want %q", tc.oids, err, tc.wantErr)
		}
	}
}

func TestObjectID_UnmarshalAmino_RoundTrip(t *testing.T) {
	pkgID := PkgIDFromPkgPath("gno.land/r/demo/x")
	for _, oid := range []ObjectID{
		// The allocator stamps PkgID before NewTime, so a stored object
		// can carry "<pkgid>:0".
		{PkgID: pkgID},
		{PkgID: pkgID, NewTime: math.MaxUint64},
	} {
		var back ObjectID
		if err := back.UnmarshalAmino(oid.String()); err != nil || back != oid {
			t.Errorf("%v: err=%v, got %v", oid, err, back)
		}
	}
}
