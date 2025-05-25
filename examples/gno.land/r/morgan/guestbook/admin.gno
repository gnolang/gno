package guestbook

import (
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/seqid"
)

var owner = ownable.NewWithOrigin()

// AdminDelete removes the guestbook message with the given ID.
// The user will still be marked as having submitted a message, so they
// won't be able to re-submit a new message.
func AdminDelete(signatureID string) {
	crossing()
	owner.AssertOwnedByPrevious()

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
