package auth

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/ed25519"
	"github.com/gnolang/gno/pkgs/crypto/multisig"
	"github.com/gnolang/gno/pkgs/crypto/secp256k1"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store"
)

var (
	// simulation signature values used to estimate gas consumption
	simSecp256k1Pubkey secp256k1.PubKeySecp256k1
	simSecp256k1Sig    [64]byte
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(simSecp256k1Pubkey[:], bz)
}

// SignatureVerificationGasConsumer is the type of function that is used to both consume gas when verifying signatures
// and also to accept or reject different types of PubKey's. This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter store.GasMeter, sig []byte, pubkey crypto.PubKey, params Params) sdk.Result

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ak AccountKeeper, bank BankKeeperI, sigGasConsumer SignatureVerificationGasConsumer) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx std.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// Get params from context.
		params := ctx.Value(AuthParamsContextKey{}).(Params)

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only ran on check tx.
		if ctx.IsCheckTx() && !simulate {
			res := EnsureSufficientMempoolFees(ctx, tx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		newCtx = SetGasMeter(simulate, ctx, tx.Fee.GasWanted)

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch ex := r.(type) {
				case store.OutOfGasException:
					log := fmt.Sprintf(
						"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
						ex.Descriptor, tx.Fee.GasWanted, newCtx.GasMeter().GasConsumed(),
					)
					res = abciResult(std.ErrOutOfGas(log))

					res.GasWanted = tx.Fee.GasWanted
					res.GasUsed = newCtx.GasMeter().GasConsumed()
					abort = true
				default:
					panic(r)
				}
			}
		}()

		if res := ValidateSigCount(tx, params); !res.IsOK() {
			return newCtx, res, true
		}

		if err := tx.ValidateBasic(); err != nil {
			return newCtx, abciResult(err), true
		}

		newCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*store.Gas(len(newCtx.TxBytes())), "txSize")

		if res := ValidateMemo(tx, params); !res.IsOK() {
			return newCtx, res, true
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := tx.GetSigners()
		signerAccs := make([]std.Account, len(signerAddrs))
		isGenesis := ctx.BlockHeight() == 0

		// fetch first signer, who's going to pay the fees
		signerAccs[0], res = GetSignerAcc(newCtx, ak, signerAddrs[0])
		if !res.IsOK() {
			return newCtx, res, true
		}

		// deduct the fees
		if !tx.Fee.GasFee.IsZero() {
			res = DeductFees(bank, newCtx, signerAccs[0], std.Coins{tx.Fee.GasFee})
			if !res.IsOK() {
				return newCtx, res, true
			}

			// reload the account as fees have been deducted
			signerAccs[0] = ak.GetAccount(newCtx, signerAccs[0].GetAddress())
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := tx.GetSignatures()

		for i := 0; i < len(stdSigs); i++ {
			// skip the fee payer, account is cached and fees were deducted already
			if i != 0 {
				signerAccs[i], res = GetSignerAcc(newCtx, ak, signerAddrs[i])
				if !res.IsOK() {
					return newCtx, res, true
				}
			}

			// check signature, return account with incremented nonce
			signBytes := GetSignBytes(newCtx.ChainID(), tx, signerAccs[i], isGenesis)
			signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytes, simulate, params, sigGasConsumer)
			if !res.IsOK() {
				return newCtx, res, true
			}

			ak.SetAccount(newCtx, signerAccs[i])
		}

		// TODO: tx tags (?)
		return newCtx, sdk.Result{GasWanted: tx.Fee.GasWanted}, false // continue...
	}
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak AccountKeeper, addr crypto.Address) (std.Account, sdk.Result) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, sdk.Result{}
	}
	return nil, abciResult(std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", addr)))
}

// ValidateSigCount validates that the transaction has a valid cumulative total
// amount of signatures.
func ValidateSigCount(tx std.Tx, params Params) sdk.Result {
	stdSigs := tx.GetSignatures()

	sigCount := 0
	for i := 0; i < len(stdSigs); i++ {
		sigCount += std.CountSubKeys(stdSigs[i].PubKey)
		if int64(sigCount) > params.TxSigLimit {
			return abciResult(std.ErrTooManySignatures(
				fmt.Sprintf("signatures: %d, limit: %d", sigCount, params.TxSigLimit),
			))
		}
	}

	return sdk.Result{}
}

// ValidateMemo validates the memo size.
func ValidateMemo(tx std.Tx, params Params) sdk.Result {
	memoLength := len(tx.GetMemo())
	if int64(memoLength) > params.MaxMemoCharacters {
		return abciResult(std.ErrMemoTooLarge(
			fmt.Sprintf(
				"maximum number of characters is %d but received %d characters",
				params.MaxMemoCharacters, memoLength,
			),
		))
	}

	return sdk.Result{}
}

// verify the signature and increment the sequence. If the account doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, acc std.Account, sig std.Signature, signBytes []byte, simulate bool, params Params,
	sigGasConsumer SignatureVerificationGasConsumer,
) (updatedAcc std.Account, res sdk.Result) {

	pubKey, res := ProcessPubKey(acc, sig, simulate)
	if !res.IsOK() {
		return nil, res
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		return nil, abciResult(std.ErrInternal("setting PubKey on signer's account"))
	}

	if simulate {
		// Simulated txs should not contain a signature and are not required to
		// contain a pubkey, so we must account for tx size of including a
		// std.Signature (Amino encoding) and simulate gas consumption
		// (assuming a SECP256k1 simulation key).
		consumeSimSigGas(ctx.GasMeter(), pubKey, sig, params)
	}

	if res := sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params); !res.IsOK() {
		return nil, res
	}

	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, abciResult(std.ErrUnauthorized("signature verification failed; verify correct account sequence and chain-id"))
	}

	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		panic(err)
	}

	return acc, res
}

func consumeSimSigGas(gasmeter store.GasMeter, pubkey crypto.PubKey, sig std.Signature, params Params) {
	simSig := std.Signature{PubKey: pubkey}
	if len(sig.Signature) == 0 {
		simSig.Signature = simSecp256k1Sig[:]
	}

	sigBz := amino.MustMarshalSized(simSig)
	cost := store.Gas(len(sigBz) + 6)

	// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
	// number of signers.
	if _, ok := pubkey.(multisig.PubKeyMultisigThreshold); ok {
		cost *= params.TxSigLimit
	}

	gasmeter.ConsumeGas(params.TxSizeCostPerByte*cost, "txSize")
}

// ProcessPubKey verifies that the given account address matches that of the
// std.Signature. In addition, it will set the public key of the account if it
// has not been set.
func ProcessPubKey(acc std.Account, sig std.Signature, simulate bool) (crypto.PubKey, sdk.Result) {
	// If pubkey is not known for account, set it from the std.Signature.
	pubKey := acc.GetPubKey()
	if simulate {
		// In simulate mode the transaction comes with no signatures, thus if the
		// account's pubkey is nil, both signature verification and gasKVStore.Set()
		// shall consume the largest amount, i.e. it takes more gas to verify
		// secp256k1 keys than ed25519 ones.
		if pubKey == nil {
			return simSecp256k1Pubkey, sdk.Result{}
		}

		return pubKey, sdk.Result{}
	}

	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, abciResult(std.ErrInvalidPubKey("PubKey not found"))
		}

		if pubKey.Address() != acc.GetAddress() {
			return nil, abciResult(std.ErrInvalidPubKey(
				fmt.Sprintf("PubKey does not match Signer address %s", acc.GetAddress())))
		}
	}

	return pubKey, sdk.Result{}
}

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter store.GasMeter, sig []byte, pubkey crypto.PubKey, params Params,
) sdk.Result {
	switch pubkey := pubkey.(type) {
	case ed25519.PubKeyEd25519:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return abciResult(std.ErrInvalidPubKey("ED25519 public keys are unsupported"))

	case secp256k1.PubKeySecp256k1:
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return sdk.Result{}

	case multisig.PubKeyMultisigThreshold:
		var multisignature multisig.Multisignature
		amino.MustUnmarshal(sig, &multisignature)

		consumeMultisignatureVerificationGas(meter, multisignature, pubkey, params)
		return sdk.Result{}

	default:
		return abciResult(std.ErrInvalidPubKey(fmt.Sprintf("unrecognized public key type: %T", pubkey)))
	}
}

func consumeMultisignatureVerificationGas(meter store.GasMeter,
	sig multisig.Multisignature, pubkey multisig.PubKeyMultisigThreshold,
	params Params) {

	size := sig.BitArray.Size()
	sigIndex := 0
	for i := 0; i < size; i++ {
		if sig.BitArray.GetIndex(i) {
			DefaultSigVerificationGasConsumer(meter, sig.Sigs[sigIndex], pubkey.PubKeys[i], params)
			sigIndex++
		}
	}
}

// DeductFees deducts fees from the given account.
//
// NOTE: We could use the CoinKeeper (in addition to the AccountKeeper, because
// the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func DeductFees(bank BankKeeperI, ctx sdk.Context, acc std.Account, fees std.Coins) sdk.Result {
	coins := acc.GetCoins()

	if !fees.IsValid() {
		return abciResult(std.ErrInsufficientFee(fmt.Sprintf("invalid fee amount: %s", fees)))
	}

	// verify the account has enough funds to pay for fees
	diff := coins.SubUnsafe(fees)
	if !diff.IsValid() {
		return abciResult(std.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds to pay for fees; %s < %s", coins, fees),
		))
	}

	err := bank.SendCoins(ctx, acc.GetAddress(), FeeCollectorAddress(), fees)
	if err != nil {
		return abciResult(err)
	}

	return sdk.Result{}
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, fee std.Fee) sdk.Result {
	minGasPrices := ctx.MinGasPrices()
	if len(minGasPrices) != 0 {

		fgw := big.NewInt(fee.GasWanted)
		fga := big.NewInt(fee.GasFee.Amount)
		fgd := fee.GasFee.Denom

		for _, gp := range minGasPrices {
			gpg := big.NewInt(gp.Gas)
			gpa := big.NewInt(gp.Price.Amount)
			gpd := gp.Price.Denom

			if fgd == gpd {
				prod1 := big.NewInt(0).Mul(fga, gpg) // fee amount * price gas
				prod2 := big.NewInt(0).Mul(fgw, gpa) // fee gas * price amount
				if prod1.Cmp(prod2) >= 0 {
					return sdk.Result{}
				} else {
					return abciResult(std.ErrInsufficientFee(
						fmt.Sprintf(
							"insufficient fees; got: %q required: %q", fee.GasFee, gp,
						),
					))
				}
			}
		}
	}

	return abciResult(std.ErrInsufficientFee(
		fmt.Sprintf(
			"insufficient fees; got: %q required (one of): %q", fee.GasFee, minGasPrices,
		),
	))
}

// SetGasMeter returns a new context with a gas meter set from a given context.
func SetGasMeter(simulate bool, ctx sdk.Context, gasLimit int64) sdk.Context {
	// In various cases such as simulation and during the genesis block, we do not
	// meter any gas utilization.
	if simulate || ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(store.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(store.NewGasMeter(gasLimit))
}

// GetSignBytes returns a slice of bytes to sign over for a given transaction
// and an account.
func GetSignBytes(chainID string, tx std.Tx, acc std.Account, genesis bool) []byte {
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}
	return std.SignBytes(
		chainID, accNum, acc.GetSequence(), tx.Fee, tx.Msgs, tx.Memo,
	)
}

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}
