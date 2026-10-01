package runtime

// PayStorage commits rlm to pay the storage deposits for its own storage, up
// to maxDeposit ugnot for the transaction.
//
// rlm is the paying realm's capability, as for PayGas. PayStorage only takes
// effect in a 0-fee sponsored transaction (which also needs a PayGas call,
// possibly by another realm), and only when rlm is the realm the message calls:
// a MsgCall to it, or its init in MsgAddPackage. Otherwise it is a no-op. It
// may be called once per transaction, and then covers rlm's own storage in
// every message of the transaction that calls rlm. Storage in other realms, or
// in messages that call another realm, stays on the caller.
//
// A deposit is refunded to whoever later frees the bytes. What rlm funded
// earlier in the same transaction is returned to rlm first; beyond that, if
// rlm lets callers free storage it sponsored, they receive its deposit.
func PayStorage(maxDeposit int64, rlm realm) {
	if !rlm.IsCurrent() {
		panic("PayStorage: rlm must be the current realm")
	}
	payStorage(rlm.PkgPath(), maxDeposit)
}

func payStorage(pkgPath string, maxDeposit int64)
