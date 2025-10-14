package ics23_test

import (
	"encoding/hex"
	"testing"

	"gno.land/p/aib/ics23"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestCommitmentProof_ExistProtoUnmarshal(t *testing.T) {
	bz, err := hex.DecodeString("0a0575676e6f7412033130301a06673175736572220a61746f6e6531757365722a046d656d6f")
	urequire.NoError(t, err)
	var p ics23.CommitmentProof_Exist

	err = p.ProtoUnmarshal(bz)

	urequire.NoError(t, err)
	urequire.NotEqual(t, nil, p.Exist)
	assertHex := func(expected string, bz []byte, msg string) {
		t.Helper()
		uassert.Equal(t,expected,hex.EncodeToString(bz),msg)
	}
	assertHex("30", p.Exist.Key,"wrong key")
	assertHex("30", p.Exist.Value,"wrong key")
	expectedValue,_:=hex.DecodeString(
	uassert.Equal(t, []byte{}, p.Exist.Value)
}
