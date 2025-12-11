package auth

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// simulation signature values used to estimate gas consumption
var simSecp256k1Pubkey secp256k1.PubKeySecp256k1

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(simSecp256k1Pubkey[:], bz)
}

// SignatureVerificationGasConsumer is the type of function that is used to both consume gas when verifying signatures
// and also to accept or reject different types of PubKey's. This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter store.GasMeter, sig []byte, pubkey crypto.PubKey, params Params) sdk.Result

type AnteOptions struct {
	// If verifyGenesisSignatures is false, does not check signatures when Height==0.
	// This is useful for development, and maybe production chains.
	// Always check your settings and inspect genesis transactions.
	VerifyGenesisSignatures bool
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ak AccountKeeper, bank BankKeeperI, sigGasConsumer SignatureVerificationGasConsumer, opts AnteOptions) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx std.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {
		// Ensure that the gas wanted is not greater than the max allowed.
		consParams := ctx.ConsensusParams()
		if consParams.Block.MaxGas == -1 {
			// no gas bounds (not recommended)
		} else if consParams.Block.MaxGas < tx.Fee.GasWanted {
			// tx gas-wanted too large.
			res = abciResult(std.ErrInvalidGasWanted(
				fmt.Sprintf(
					"invalid gas-wanted; got: %d block-max-gas: %d",
					tx.Fee.GasWanted, consParams.Block.MaxGas,
				),
			))
			return ctx, res, true
		}

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only run upon checktx.
		if ctx.IsCheckTx() && !simulate {
			res := EnsureSufficientMempoolFees(ctx, tx.Fee)
			if !res.IsOK() {
				return ctx, res, true
			}
		}

		newCtx = SetGasMeter(ctx, tx.Fee.GasWanted)

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch ex := r.(type) {
				case store.OutOfGasError:
					log := fmt.Sprintf(
						"out of gas in location: %v; gasWanted: %d, gasUsed: %d (until panic)",
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

		// Get params from context.
		params := ctx.Value(AuthParamsContextKey{}).(Params)
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
			res = DeductFees(bank, newCtx, signerAccs[0], ak.FeeCollectorAddress(ctx), std.Coins{tx.Fee.GasFee})
			if !res.IsOK() {
				return newCtx, res, true
			}

			// reload the account as fees have been deducted
			signerAccs[0] = ak.GetAccount(newCtx, signerAccs[0].GetAddress())
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := tx.GetSignatures()

		for i := range stdSigs {
			// skip the fee payer, account is cached and fees were deducted already
			if i != 0 {
				signerAccs[i], res = GetSignerAcc(newCtx, ak, signerAddrs[i])
				if !res.IsOK() {
					return newCtx, res, true
				}
			}

			// check signature, return account with incremented nonce
			sacc := signerAccs[i]
			if isGenesis && !opts.VerifyGenesisSignatures {
				// No signatures are needed for genesis.
			} else {
				// Check signature
				signBytes, err := GetSignBytes(newCtx.ChainID(), tx, sacc, isGenesis)
				if err != nil {
					return newCtx, res, true
				}
				signerAccs[i], res = processSig(newCtx, sacc, stdSigs[i], signBytes, simulate, params, sigGasConsumer)
				if !res.IsOK() {
					return newCtx, res, true
				}
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
	for i := range stdSigs {
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
	if int64(memoLength) > params.MaxMemoBytes {
		return abciResult(std.ErrMemoTooLarge(
			fmt.Sprintf(
				"maximum number of bytes is %d but received %d bytes",
				params.MaxMemoBytes, memoLength,
			),
		))
	}

	return sdk.Result{}
}

// verify the signature and increment the sequence. If the account doesn't
// have a pubkey, set it.
func processSig(
	ctx sdk.Context, acc std.Account, sig std.Signature, signBytes []byte, simulate bool, params Params,
	sigGasConsumer SignatureVerificationGasConsumer,
) (updatedAcc std.Account, res sdk.Result) {
	pubKey, res := ProcessPubKey(acc, sig)
	if !res.IsOK() {
		return nil, res
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		return nil, abciResult(std.ErrInternal("setting PubKey on signer's account"))
	}

	if res := sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params); !res.IsOK() {
		return nil, res
	}

	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, abciResult(std.ErrUnauthorized("signature verification failed; verify correct account, sequence, and chain-id"))
	}

	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		panic(err)
	}

	return acc, res
}

// ProcessPubKey verifies that the given account address matches that of the
// std.Signature. In addition, it will set the public key of the account if it
// has not been set.
func ProcessPubKey(acc std.Account, sig std.Signature) (crypto.PubKey, sdk.Result) {
	// If pubkey is not known for account, set it from the std.Signature.
	pubKey := acc.GetPubKey()
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

// DefaultSigVerificationGasConsumer is the default implementation of
// SignatureVerificationGasConsumer. It consumes gas for signature verification
// based upon the public key type. The cost is fetched from the given params
// and is matched by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter store.GasMeter, sig []byte, pubkey crypto.PubKey, params Params,
) sdk.Result {
	switch pubkey := pubkey.(type) {
	case ed25519.PubKeyEd25519:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return sdk.Result{}

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
	params Params,
) {
	size := sig.BitArray.Size()
	sigIndex := 0
	for i := range size {
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
func DeductFees(bk BankKeeperI, ctx sdk.Context, acc std.Account, collector crypto.Address, fees std.Coins) sdk.Result {
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

	// Sending coins is unrestricted to pay for gas fees
	err := bk.SendCoinsUnrestricted(ctx, acc.GetAddress(), collector, fees)
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
	blockGasPrice := ctx.Value(GasPriceContextKey{}).(std.GasPrice)
	feeGasPrice := std.GasPrice{
		Gas: fee.GasWanted,
		Price: std.Coin{
			Amount: fee.GasFee.Amount,
			Denom:  fee.GasFee.Denom,
		},
	}
	// check the block gas price
	if blockGasPrice.Price.IsValid() && !blockGasPrice.Price.IsZero() {
		ok, err := feeGasPrice.IsGTE(blockGasPrice)
		if err != nil {
			return abciResult(std.ErrInsufficientFee(
				err.Error(),
			))
		}
		if !ok {
			return abciResult(std.ErrInsufficientFee(
				fmt.Sprintf(
					"insufficient fees; got: {Gas-Wanted: %d, Gas-Fee %s}, fee required: %+v as block gas price", feeGasPrice.Gas, feeGasPrice.Price, blockGasPrice,
				),
			))
		}
	}
	// check min gas price set by the node.
	if len(minGasPrices) == 0 {
		// no minimum gas price (not recommended)
		// TODO: allow for selective filtering of 0 fee txs.
		return sdk.Result{}
	} else {
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
				// This is equivalent to checking
				// That the Fee / GasWanted ratio is greater than or equal to the minimum GasPrice per gas.
				// This approach helps us avoid dealing with configurations where the value of
				// the minimum gas price is set to 0.00001ugnot/gas.
				if prod1.Cmp(prod2) >= 0 {
					return sdk.Result{}
				} else {
					fee := new(big.Int).Quo(prod2, gpg)
					return abciResult(std.ErrInsufficientFee(
						fmt.Sprintf(
							"insufficient fees; got: {Gas-Wanted: %d, Gas-Fee %s}, fee required: %d with %+v as minimum gas price set by the node", feeGasPrice.Gas, feeGasPrice.Price, fee, gp,
						),
					))
				}
			}
		}
	}

	return abciResult(std.ErrInsufficientFee(
		fmt.Sprintf(
			"insufficient fees; got: {Gas-Wanted: %d, Gas-Fee %s}, required (one of): %q", feeGasPrice.Gas, feeGasPrice.Price, minGasPrices,
		),
	))
}

// SetGasMeter returns a new context with a gas meter set from a given context.
func SetGasMeter(ctx sdk.Context, gasLimit int64) sdk.Context {
	if ctx.Mode() == sdk.RunTxModeSimulate {
		// Cap simulation gas to avoid unbounded consumption; use consensus maxGas
		// as a ceiling. Ignore the tx gasLimit here so we can measure full gas usage.
		maxGas := int64(-1)
		if cp := ctx.ConsensusParams(); cp != nil {
			maxGas = cp.Block.MaxGas
		}

		if maxGas != int64(-1) {
			return ctx.WithGasMeter(store.NewGasMeter(maxGas))
		}

		// If no limit is configured (maxGas == -1), fall back to infinite.
		return ctx.WithGasMeter(store.NewInfiniteGasMeter())
	}

	// In various cases such as simulation and during the genesis block, we do not
	// meter any gas utilization.
	if ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(store.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(store.NewGasMeter(gasLimit))
}

// GetSignBytes returns a slice of bytes to sign over for a given transaction
// and an account.
func GetSignBytes(chainID string, tx std.Tx, acc std.Account, genesis bool) ([]byte, error) {
	var (
		accNum      uint64
		accSequence uint64
	)
	if !genesis {
		accNum = acc.GetAccountNumber()
		accSequence = acc.GetSequence()
	}

	return std.GetSignaturePayload(
		std.SignDoc{
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      accSequence,
			Fee:           tx.Fee,
			Msgs:          tx.Msgs,
			Memo:          tx.Memo,
		},
	)
}

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}
