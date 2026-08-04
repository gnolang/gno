package runtime

// PayGas commits the calling realm to pay gas for the current transaction.
// maxFee is the maximum amount in ugnot the realm is willing to pay.
// The gas limit is derived from maxFee and the current (dynamic) gas price —
// the same price used at settlement — not the per-validator minimum gas price.
//
// Rules:
//   - Only realms can call PayGas (not packages, not main).
//   - The function containing PayGas must be defined in the same realm that pays.
//   - Can only be called once per transaction.
//   - Only takes effect in a 0-fee sponsored transaction (a no-op otherwise).
//
// PayGas and PayStorage are independent: one realm may sponsor gas while a
// different realm sponsors storage in the same transaction. Each realm pays only
// its own commitment from its own balance.
func PayGas(maxFee int64)
