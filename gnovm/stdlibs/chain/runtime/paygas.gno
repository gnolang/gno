package runtime

// PayGas commits the calling realm to pay gas for the current transaction.
// maxFee is the maximum amount in ugnot the realm is willing to pay.
// The gas limit is derived from maxFee / MinGasPrice.
//
// Rules:
//   - Only realms can call PayGas (not packages, not main).
//   - The function containing PayGas must be defined in the same realm that pays.
//   - Can only be called once per transaction.
func PayGas(maxFee int64)
