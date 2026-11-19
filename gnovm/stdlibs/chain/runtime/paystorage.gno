package runtime

// PayStorage commits rlm to pay the storage deposits for its own storage, up
// to maxDeposit ugnot for the transaction.
//
// rlm is the paying realm's capability, as for PayGas, and invalid arguments
// always panic. A valid call only takes effect in a 0-fee sponsored transaction
// (which also needs a PayGas call, possibly by another realm), and only when rlm
// is the realm the message calls: a MsgCall to it, or its init in
// MsgAddPackage. Otherwise it is a no-op. It may be called once per
// transaction, and then covers rlm's storage in every message of the
// transaction that calls rlm. Storage in other realms, or in messages that call
// another realm, stays on the caller.
//
// Everything those messages add to rlm's storage is charged to rlm, whoever's
// code adds it, for example objects rlm allocates for another realm to keep. A
// deposit is refunded to whoever later frees the bytes, so a sponsored function
// should not run code it does not control, nor let callers free what it
// sponsored. Within one transaction, a free in rlm's storage is first netted
// against what rlm locked earlier in it, whoever paid for the freed bytes.
func PayStorage(maxDeposit int64, rlm realm) {
	if !rlm.IsCurrent() {
		panic("PayStorage: rlm must be the current realm")
	}
	payStorage(rlm.PkgPath(), maxDeposit)
}

func payStorage(pkgPath string, maxDeposit int64)
