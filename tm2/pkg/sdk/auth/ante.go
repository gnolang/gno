package auth

import (
	"bytes"
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

		signerAddrs := tx.GetSigners()
		signerAccs := make([]std.Account, len(signerAddrs))
		stdSigs := tx.GetSignatures()
		isGenesis := ctx.BlockHeight() == 0
		sessionAccounts := map[crypto.Address]std.DelegatedAccount{}

		// ——— Phase 1: Resolve all signers ———

		for i, signerAddr := range signerAddrs {
			signerAccs[i], res = GetSignerAcc(newCtx, ak, signerAddr)
			if !res.IsOK() {
				return newCtx, res, true
			}

			if !stdSigs[i].SessionAddr.IsZero() {
				sa := ak.GetSessionAccount(newCtx, signerAddr, stdSigs[i].SessionAddr)
				if sa == nil {
					return newCtx, abciResult(std.ErrUnauthorized("unknown session")), true
				}
				da := sa.(std.DelegatedAccount)
				if da.GetExpiresAt() > 0 && newCtx.BlockTime().Unix() >= da.GetExpiresAt() {
					return newCtx, abciResult(std.ErrSessionExpired(fmt.Sprintf(
						"session expired: expires_at=%d, block_time=%d",
						da.GetExpiresAt(), newCtx.BlockTime().Unix()))), true
				}
				sessionAccounts[signerAddr] = da
			}
		}

		// ——— Phase 2: Pre-check session outflow, then deduct gas fees ———

		// Phase 2a: If the first signer is a session, pre-check its total
		// declared outflow (gas fee + each msg's SpendForSigner) against
		// the session's remaining SpendLimit BEFORE any deduction. This
		// rejects obviously-over-limit session-signed txs without charging
		// gas, preventing a mempool-gas-bleed attack where a compromised
		// session could submit many doomed txs and bleed gas from master
		// on each ante Phase 2 commit.
		//
		// Msgs that don't implement std.SpendEstimator are skipped here;
		// the bank.Keeper.SendCoins session hook still catches their
		// actual outflow at execution time, so correctness is unchanged —
		// this pre-check is purely a gas-efficiency optimization.
		if da, ok := sessionAccounts[signerAddrs[0]]; ok {
			total := std.Coins{}
			if !tx.Fee.GasFee.IsZero() {
				total = total.Add(std.Coins{tx.Fee.GasFee})
			}
			for _, msg := range tx.GetMsgs() {
				if est, ok := msg.(std.SpendEstimator); ok {
					total = total.Add(est.SpendForSigner(signerAddrs[0]))
				}
			}
			if err := CheckSessionSpend(da, total, newCtx.BlockTime().Unix()); err != nil {
				return newCtx, abciResult(err), true
			}
		}

		// Phase 2b: Deduct gas fees from first signer (always master).
		if !tx.Fee.GasFee.IsZero() {
			// Gas fees count against session spend limits.
			if da, ok := sessionAccounts[signerAddrs[0]]; ok {
				if err := DeductSessionSpend(da, std.Coins{tx.Fee.GasFee}, newCtx.BlockTime().Unix()); err != nil {
					return newCtx, abciResult(err), true
				}
				// SpendUsed updated on in-memory da; persisted in Phase 3.
			}
			res = DeductFees(bank, newCtx, signerAccs[0], ak.FeeCollectorAddress(ctx), std.Coins{tx.Fee.GasFee})
			if !res.IsOK() {
				return newCtx, res, true
			}
			// reload the account as fees have been deducted
			signerAccs[0] = ak.GetAccount(newCtx, signerAddrs[0])
		}

		// ——— Phase 3: Verify signatures, increment sequences ———

		for i, sig := range stdSigs {
			if isGenesis && !opts.VerifyGenesisSignatures {
				continue
			}

			da, isSession := sessionAccounts[signerAddrs[i]]

			// Pick the account that holds the pubkey + sequence.
			var sigAcc std.Account
			if isSession {
				sigAcc = da.(std.Account)
			} else {
				sigAcc = signerAccs[i]
			}

			// Resolve pubkey.
			pubKey := sig.PubKey
			if pubKey == nil {
				// No pubkey in signature — use stored key.
				pubKey = sigAcc.GetPubKey()
			} else if sigAcc.GetPubKey() == nil {
				// First tx: set pubkey on account.
				//
				// Asymmetry between master and session accounts is intentional.
				// For MASTER accounts, we MUST verify that the supplied pubkey
				// hashes to the signer address, because master addresses are
				// derived lazily on first interaction — the first signer to
				// claim a never-seen address can fix its pubkey, so we must
				// reject an address-mismatched pubkey to prevent pubkey squats.
				//
				// For SESSION accounts, the address was set at CREATION time
				// via keeper.NewSessionAccount using msg.SessionKey.Address()
				// (see auth/keeper.go:NewSessionAccount). The handler already
				// enforced that sessionAddr == msg.SessionKey.Address() and
				// rejected collisions with existing accounts. So by the time
				// we reach this branch for a session, sigAcc.GetAddress() is
				// guaranteed to equal the pubkey's derived address — there's
				// nothing to verify.
				if !isSession {
					// For master accounts, verify pubkey matches address.
					if pubKey.Address() != sigAcc.GetAddress() {
						return newCtx, abciResult(std.ErrInvalidPubKey(
							fmt.Sprintf("PubKey does not match Signer address %s", sigAcc.GetAddress()))), true
					}
				}
				sigAcc.SetPubKey(pubKey)
			} else {
				// Both sig.PubKey and stored pubkey exist — they must match.
				if !bytes.Equal(pubKey.Bytes(), sigAcc.GetPubKey().Bytes()) {
					return newCtx, abciResult(std.ErrUnauthorized("signature verification failed; verify correct account, sequence, and chain-id")), true
				}
				pubKey = sigAcc.GetPubKey()
			}
			if pubKey == nil {
				return newCtx, abciResult(std.ErrInvalidPubKey("PubKey not found")), true
			}

			// Sign bytes: sigAcc's own AccountNumber and Sequence.
			// At genesis, both are zero regardless of actual values.
			var accNum, accSeq uint64
			if !isGenesis {
				accNum = sigAcc.GetAccountNumber()
				accSeq = sigAcc.GetSequence()
			}
			signBytes, err := tx.GetSignBytes(
				newCtx.ChainID(),
				accNum,
				accSeq,
			)
			if err != nil {
				return newCtx, abciResult(std.ErrInternal("getting sign bytes")), true
			}

			if res := sigGasConsumer(newCtx.GasMeter(), sig.Signature, pubKey, params); !res.IsOK() {
				return newCtx, res, true
			}

			if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
				return newCtx, abciResult(std.ErrUnauthorized("signature verification failed; verify correct account, sequence, and chain-id")), true
			}

			if isSession {
				sigAcc.SetSequence(sigAcc.GetSequence() + 1)
				ak.SetSessionAccount(newCtx, signerAddrs[i], sigAcc)
			} else {
				sigAcc.SetSequence(sigAcc.GetSequence() + 1)
				ak.SetAccount(newCtx, signerAccs[i])
			}
		}

		// ——— Phase 4: Propagate session accounts in context ———

		if len(sessionAccounts) > 0 {
			newCtx = newCtx.WithValue(std.SessionAccountsContextKey{}, sessionAccounts)
		}

		return newCtx, sdk.Result{GasWanted: tx.Fee.GasWanted}, false
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
