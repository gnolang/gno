package runtime

// PayStorage commits the calling realm to pay storage deposits for the current transaction.
// maxDeposit is the maximum amount in ugnot the realm is willing to pay for storage.
// If total storage deposits exceed maxDeposit, the transaction fails.
//
// Same rules as PayGas:
//   - Only realms can call PayStorage (not packages, not main).
//   - The function containing PayStorage must be defined in the same realm that pays.
//   - Can only be called once per transaction.
//   - Only takes effect in a 0-fee sponsored transaction (a no-op otherwise).
//
// PayGas and PayStorage are independent: one realm may sponsor gas while a
// different realm sponsors storage in the same transaction. Each realm pays only
// its own commitment from its own balance.
func PayStorage(maxDeposit int64)
