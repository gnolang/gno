package runtime

// PayStorage commits the calling realm to pay storage deposits for the current transaction.
// maxDeposit is the maximum amount in ugnot the realm is willing to pay for storage.
// If total storage deposits exceed maxDeposit, the transaction fails.
//
// Same rules as PayGas:
//   - Only realms can call PayStorage (not packages, not main).
//   - The function containing PayStorage must be defined in the same realm that pays.
//   - Can only be called once per transaction.
//   - Must be called by the same realm as PayGas (if both are used).
func PayStorage(maxDeposit int64)
