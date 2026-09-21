package guestbook

import (
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/seqid/v0"
)

var owner *ownable.Ownable

func init(cur realm) {
	if !cur.Previous().IsUserCall() {
		panic("realm must be initialized by EOA")
	}
	owner = ownable.NewWithAddress(cur.Previous().Address())
}

// AdminDelete removes the guestbook message with the given ID.
// The user will still be marked as having submitted a message, so they
// won't be able to re-submit a new message.
func AdminDelete(cur realm, signatureID string) {
	owner.AssertOwnedBy(cur.Previous().Address())
	id, err := seqid.FromString(signatureID)
	if err != nil {
		panic(err)
	}
	idb := id.Binary()
	if !guestbook.Has(idb) {
		panic("signature does not exist")
	}
	guestbook.Remove(idb)
}
