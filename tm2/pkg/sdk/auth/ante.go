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
	"github.com/gnolang/gno/tm2/pkg/overflow"
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
	// AllowZeroFeeTxs enables 0-fee transactions when realms sponsor gas via PayGas.
	AllowZeroFeeTxs bool

	// RequireSigForSimulate reports whether tx must have its signatures
	// cryptographically verified even in simulate mode.
	//
	// Simulate normally skips verification so a caller can estimate gas
	// without holding a key. That is safe only for messages whose
	// authorization does not depend on who signed: `.app/simulate` is a
	// public query that executes the messages, so for a message authorized
	// by signer identity the skip would let an unauthenticated caller name
	// somebody else's address and have it accepted. Applications set this
	// for those message types.
	//
	// A missing or miscounted signature is already refused for every mode by
	// tx.ValidateBasic below, so this only closes the verification gap.
	//
	// Note for callers that estimate gas without signing: a tx containing a
	// message selected by this predicate must carry a real signature, not a
	// pubkey-only placeholder.
	RequireSigForSimulate func(tx std.Tx) bool
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(ak AccountKeeper, bank BankKeeperI, sigGasConsumer SignatureVerificationGasConsumer, opts AnteOptions) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx std.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {
		// Determine if this is a 0-fee PayGas transaction.
		consParams := ctx.ConsensusParams()
		isZeroFeeTx := tx.Fee.GasFee.IsZero() && consParams.Block.MaxGasCreditPerTx > 0

		// Genesis is exempt from the two SponsorStorage guards below — genesis txs
		// are trusted and never sponsor. "Genesis" is a property of the DELIVERY,
		// not of the raw height: CheckTx is served from node start and checkState
		// carries InitChain's zero height until the first Commit, so gating on
		// height alone would let ordinary mempool traffic skip both guards in that
		// window and admit txs that are then rejected deterministically at
		// DeliverTx. Mirrors the credit-window scoping in baseapp.runTx.
		notGenesis := ctx.BlockHeight() > 0 || ctx.Mode() != sdk.RunTxModeDeliver

		// A sponsored (0-fee) tx cannot be ADMITTED before the first block is
		// committed. checkState carries InitChain's zero height until the first
		// Commit, and at height 0 gno.land's genesis wrapper auto-creates and
		// funds any unknown signer (see gnoland.NewAppWithOptions). Admission
		// decisions taken against that synthetic state are meaningless: a tx from
		// a fresh key passes CheckTx and is gossiped, then fails deterministically
		// at DeliverTx in block 1 where the account never existed — and repeating
		// it with fresh keys forces one credit-window VM execution per attempt on
		// every opting-in node, during startup.
		//
		// Genesis DELIVERY is unaffected (trusted genesis txs never sponsor), and
		// there is nothing to lose by refusing: the chain is not producing blocks
		// yet, so clients simply resubmit once block 1 exists.
		if isZeroFeeTx && ctx.BlockHeight() == 0 && ctx.Mode() != sdk.RunTxModeDeliver {
			res = abciResult(std.ErrUnauthorized(
				"sponsored transactions are not accepted before the first block"))
			return ctx, res, true
		}

		// A 0-fee tx is only legitimate when the credit window is OPEN, because
		// that is the only path that gives it a payer. With the window disabled
		// (MaxGasCreditPerTx == 0 — the default, and the kill switch) there is no
		// sponsorship: PayGas is inert, no settlement runs, and the tx would
		// simply execute for free on the caller's GasWanted.
		//
		// This must be rejected in EVERY mode, not just CheckTx. The mempool
		// minimum-fee check below is deliberately CheckTx-only (it is local
		// validator policy), so it cannot stop a proposer from force-including a
		// 0-fee tx. Before this feature, Tx.ValidateBasic rejected such txs
		// outright; relaxing it to admit canonical zero fees for sponsorship
		// removed that backstop, and this restores it.
		if tx.Fee.GasFee.IsZero() && !isZeroFeeTx && notGenesis {
			res = abciResult(std.ErrInsufficientFee(
				"zero-fee transactions require a non-zero Block.MaxGasCreditPerTx"))
			return ctx, res, true
		}

		// Fee.SponsorStorage only applies to sponsored (0-fee) txs, where a realm
		// covers deferred storage via PayStorage. Reject it on a normal fee-paying
		// tx so the mistake surfaces at submission (CheckTx) rather than failing
		// opaquely at inclusion (the deferred path would skip the signer's
		// per-message deposit and then abort at end-of-tx).
		if tx.Fee.SponsorStorage && !isZeroFeeTx && notGenesis {
			res = abciResult(std.ErrUnauthorized("SponsorStorage requires a 0-fee sponsored transaction"))
			return ctx, res, true
		}

		// SponsorStorage defers all messages' storage diffs to end-of-tx, where
		// the per-message caller identity is lost: the deferred settlement can
		// attribute a freed-storage refund only to a single tx caller (the first
		// signer). Restrict it to single-signer txs so that caller is unambiguous,
		// avoiding routing one signer's refund to a co-signer.
		if tx.Fee.SponsorStorage && len(tx.GetSigners()) > 1 && notGenesis {
			res = abciResult(std.ErrUnauthorized("SponsorStorage is not supported for multi-signer transactions"))
			return ctx, res, true
		}

		// Ensure that the gas wanted is not greater than the max allowed.
		// For 0-fee txs, gas limit is set by the credit window, not GasWanted.
		if !isZeroFeeTx {
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
		}

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only run upon checktx. Skip for 0-fee PayGas txs when allowed.
		if ctx.IsCheckTx() && !simulate {
			if isZeroFeeTx && !opts.AllowZeroFeeTxs {
				res = abciResult(std.ErrInsufficientFee("zero-fee transactions not accepted by this validator"))
				return ctx, res, true
			}
			if !isZeroFeeTx {
				res := EnsureSufficientMempoolFees(ctx, tx.Fee)
				if !res.IsOK() {
					return ctx, res, true
				}
			}
		}

		// Set gas meter: credit window for 0-fee txs, GasWanted for normal txs.
		if isZeroFeeTx {
			newCtx = SetGasMeter(ctx, consParams.Block.MaxGasCreditPerTx)
			// SetGasMeter hands out an INFINITE meter at height 0 (genesis) and
			// for source-gas replay. That exemption is correct when DELIVERING a
			// trusted genesis/replay tx, but CheckTx is served from node start
			// while checkState still carries InitChain's zero height, so mempool
			// admission of a 0-fee tx would otherwise run the VM unmetered.
			// Outside DeliverTx, always bound a credit-window tx by the window.
			if ctx.Mode() != sdk.RunTxModeDeliver {
				newCtx = newCtx.WithGasMeter(store.NewGasMeter(consParams.Block.MaxGasCreditPerTx))
			}
		} else {
			newCtx = SetGasMeter(ctx, tx.Fee.GasWanted)
		}

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch ex := r.(type) {
				case store.OutOfGasError:
					gasUsed := newCtx.GasMeter().GasConsumed()
					maxGas := int64(-1)
					if cp := newCtx.ConsensusParams(); cp != nil && cp.Block != nil {
						maxGas = cp.Block.MaxGas
					}
					log := store.OutOfGasLog(gasUsed, tx.Fee.GasWanted, maxGas, ex.Descriptor, true)
					res = abciResult(std.ErrOutOfGas(log))

					res.GasWanted = tx.Fee.GasWanted
					res.GasUsed = gasUsed
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

		// Mulp, not a bare multiply: TxSizeCostPerByte is only validated positive,
		// so an absurd one wraps. Some wraps land negative and the meter refuses
		// them, but others land small and positive -- a 4096-byte transaction
		// charged 4096 gas for its size, silently. Panic on the overflow instead,
		// as the same per-byte multiply does in gno.land/pkg/sdk/vm.
		newCtx.GasMeter().ConsumeGas(
			overflow.Mulp(params.TxSizeCostPerByte, store.Gas(len(newCtx.TxBytes()))), "txSize")

		if res := ValidateMemo(tx, params); !res.IsOK() {
			return newCtx, res, true
		}

		signerAddrs := tx.GetSigners()
		signerAccs := make([]std.Account, len(signerAddrs))
		stdSigs := tx.GetSignatures()
		isGenesis := ctx.BlockHeight() == 0
		sessionAccounts := map[crypto.Address]std.DelegatedAccount{}

		// Store tx caller and sponsor flag for end-of-tx settlement.
		newCtx = newCtx.WithTxCaller(signerAddrs[0]).WithSponsorStorage(tx.Fee.SponsorStorage)

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
			// Hardfork genesis replay: historical and patched txs carry a
			// BlockHeight > 0 overridden for faithful re-execution, so the
			// isGenesis check above misses them. When the operator opted
			// into --skip-genesis-sig-verification, skip their signature
			// check too — the whole replayed genesis is vouched for by its
			// agreed sha256, and a rewritten (patched) body can no longer
			// verify by design. isGenesis is left untouched so the
			// accNum/accSeq sign-bytes logic below still uses source values.
			if !opts.VerifyGenesisSignatures {
				if IsGenesisReplay(ctx) {
					continue
				}
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

			// Simulate normally skips verification; see RequireSigForSimulate
			// for why some messages cannot afford that.
			//
			// CheckTx admission of 0-fee sponsored txs runs in
			// RunTxModeCheckExecute, not Simulate, so simulate is false there
			// and forged-signature txs are rejected before entering the mempool.
			verifySig := !simulate ||
				(opts.RequireSigForSimulate != nil && opts.RequireSigForSimulate(tx))
			if verifySig && !pubKey.VerifyBytes(signBytes, sig.Signature) {
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

		// Report GasWanted. For 0-fee txs the effective per-tx gas ceiling is the
		// credit window (the meter was sized to MaxGasCreditPerTx above), NOT the
		// client-supplied tx.Fee.GasWanted. The mempool sums the reported GasWanted
		// against Block.MaxGas when packing a block, so reporting the credit window
		// keeps block packing bounded by real worst-case consumption; reporting the
		// client value (which can be 0) would let a proposer overfill the block.
		reportedGasWanted := tx.Fee.GasWanted
		if isZeroFeeTx {
			reportedGasWanted = consParams.Block.MaxGasCreditPerTx
		}
		return newCtx, sdk.Result{GasWanted: reportedGasWanted}, false
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
		// Bound the depth and the total size of the recursion below before
		// decoding anything: the key alone decides both, and each level costs a
		// decode of whatever signature bytes remain. ValidateSigCount bounds
		// neither, because std.CountSubKeys counts leaves: a chain of 1-of-1
		// keys has exactly one however deep it runs, and a constituent holding
		// no keys at all has none however many of them are listed.
		if err := pubkey.ValidateStructure(); err != nil {
			return abciResult(std.ErrInvalidPubKey(err.Error()))
		}

		var multisignature multisig.Multisignature
		// sig is the signature field of an untrusted transaction, so this must
		// not be MustUnmarshal: arbitrary bytes there would panic out of the
		// ante handler, whose own recover only handles OutOfGasError, and be
		// caught by runTx's blanket recover as an ErrInternal with a stack
		// trace. The amino error is not reported back because it renders the
		// offending buffer as hex, and the buffer is caller-sized.
		if err := amino.Unmarshal(sig, &multisignature); err != nil {
			return abciResult(std.ErrUnauthorized("signature is not a valid multisignature"))
		}

		return consumeMultisignatureVerificationGas(meter, multisignature, pubkey, params)

	default:
		return abciResult(std.ErrInvalidPubKey(fmt.Sprintf("unrecognized public key type: %T", pubkey)))
	}
}

// consumeMultisignatureVerificationGas consumes gas for each signed subkey of
// pubkey.
//
// The returned result MUST be checked by the caller: it is what rejects subkeys
// whose type is not recognized as a signing key type. Dropping it would let a
// key type that VerifyBytes handles but this function does not (e.g. a mock key
// with trivially forgeable signatures) reach PubKeyMultisigThreshold.VerifyBytes
// as a constituent key, forging the multisig signature as a whole.
func consumeMultisignatureVerificationGas(meter store.GasMeter,
	sig multisig.Multisignature, pubkey multisig.PubKeyMultisigThreshold,
	params Params,
) sdk.Result {
	// Establish the shape of the signature against the key before walking it: a
	// bit array whose ExtraBitsStored runs past the end of its Elems decodes
	// perfectly well, so the amino error above does not catch it, and then makes
	// Size() report more bits than are stored — which the walk below indexes.
	// One implementation, shared with PubKeyMultisigThreshold.VerifyBytes, so
	// that the two cannot come to disagree about which shapes are walkable.
	if err := sig.ValidateBasic(len(pubkey.PubKeys)); err != nil {
		return abciResult(std.ErrUnauthorized(err.Error()))
	}

	size := sig.BitArray.Size()
	sigIndex := 0
	for i := range size {
		if sig.BitArray.GetIndex(i) {
			if res := DefaultSigVerificationGasConsumer(meter, sig.Sigs[sigIndex], pubkey.PubKeys[i], params); !res.IsOK() {
				return res
			}
			sigIndex++
		}
	}
	return sdk.Result{}
}

// DeductFees deducts fees from the given account.
//
// NOTE: We could use the CoinKeeper (in addition to the AccountKeeper, because
// the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func DeductFees(bk BankKeeperI, ctx sdk.Context, acc std.Account, collector crypto.Address, fees std.Coins) sdk.Result {
	if !fees.IsValid() {
		return abciResult(std.ErrInsufficientFee(fmt.Sprintf("invalid fee amount: %s", fees)))
	}

	// Verify the account has enough funds to pay for fees, one fee denom at a
	// time. Read through the bank rather than acc.GetCoins(): a balance does not
	// necessarily live in the account object, since realm-issued denoms have
	// their own keys, and the fee denom is whatever the transaction names.
	// Reading only the fee denoms also keeps this independent of how many other
	// denoms the payer happens to hold.
	addr := acc.GetAddress()
	for _, fee := range fees {
		if balance := bk.GetCoin(ctx, addr, fee.Denom); balance < fee.Amount {
			return abciResult(std.ErrInsufficientFunds(
				fmt.Sprintf("insufficient funds to pay for fees; %d%s < %s", balance, fee.Denom, fee),
			))
		}
	}

	// Sending coins is unrestricted to pay for gas fees
	err := bk.SendCoinsUnrestricted(ctx, addr, collector, fees)
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
		fgd := fee.GasFee.Denom

		for _, gp := range minGasPrices {
			if fgd != gp.Price.Denom {
				continue
			}
			// Decided by GasPrice.IsGTE, the same comparison the block minimum
			// above uses, so one implementation owns the rule. Cross-multiplying
			// here instead meant this copy did not inherit IsGTE's guards: a
			// negative gas_wanted flips the sign of one side, and a fee of nothing
			// then compares as sufficient.
			ok, err := feeGasPrice.IsGTE(gp)
			if err != nil {
				return abciResult(std.ErrInsufficientFee(err.Error()))
			}
			if ok {
				return sdk.Result{}
			}
			// What the fee should have been, for the message. ParseGasPrice
			// refuses a non-positive gp.Gas and IsGTE has already refused a
			// non-positive gas_wanted, so this cannot divide by zero.
			required := new(big.Int).Quo(
				new(big.Int).Mul(fgw, big.NewInt(gp.Price.Amount)),
				big.NewInt(gp.Gas),
			)
			return abciResult(std.ErrInsufficientFee(
				fmt.Sprintf(
					"insufficient fees; got: {Gas-Wanted: %d, Gas-Fee %s}, fee required: %d with %+v as minimum gas price set by the node", feeGasPrice.Gas, feeGasPrice.Price, required, gp,
				),
			))
		}
	}

	return abciResult(std.ErrInsufficientFee(
		fmt.Sprintf(
			"insufficient fees; got: {Gas-Wanted: %d, Gas-Fee %s}, required (one of): %q", feeGasPrice.Gas, feeGasPrice.Price, minGasPrices,
		),
	))
}

// SkipGasMeteringKey is a context key used to bypass gas metering for
// historical tx replay during chain upgrades. When set on the context,
// SetGasMeter installs an infinite gas meter even for non-genesis blocks.
// Used by gnoland's GasReplayMode="source" during genesis replay to
// preserve source-chain outcomes when gas requirements have changed.
type SkipGasMeteringKey struct{}

// GenesisReplayKey is a context key marking a tx delivery as part of InitChain.
// It is set for EVERY such tx, including a fresh chain's own genesis txs -- not
// only history replayed from a previous chain. The hardfork case is why it
// exists: replayed txs carry a BlockHeight > 0 (overridden for faithful
// re-execution), so a ctx.BlockHeight()==0 check under-reports them. But
// callers depend on the broad reading too, so do not narrow it to the hardfork
// case. See IsGenesisReplay.
//
// It never bypasses signature verification on its own: the ante skips
// verification for a replay tx only when the node was also started with
// --skip-genesis-sig-verification (VerifyGenesisSignatures=false). In a
// normally-configured node that flag is unset, so this key has no effect
// on signature verification. Set only by gnoland's InitChainer per-tx
// delivery wrapper.
type GenesisReplayKey struct{}

// IsGenesisReplay reports whether ctx is delivering a transaction as part of
// InitChain, rather than live traffic.
//
// Note the wording: it is true for EVERY tx delivered during InitChain, which
// includes a fresh chain's own genesis txs, not only history replayed from a
// previous chain. gnoland's delivery wrapper sets it unconditionally. Callers
// that want "this is replayed history specifically" must also test the tx's
// metadata; nothing here distinguishes the two, and a fresh launch relies on
// the broad reading -- gno.land's code-submission gate exempts genesis MsgRun
// through exactly this predicate, which is how a chain seeds its first DAO
// members before any allowlist exists.
//
// The predicate belongs beside the key: four places now branch on it — this
// package's ante, gno.land's code-submission gate, and the vm keeper's inert
// and enable paths — and each open-coded `ctx.Value(...).(bool)` is a chance to
// get the key type or the comma-ok wrong silently, in a direction that fails
// open.
func IsGenesisReplay(ctx sdk.Context) bool {
	replay, _ := ctx.Value(GenesisReplayKey{}).(bool)
	return replay
}

// SetGasMeter returns a new context with a gas meter set from a given context.
func SetGasMeter(ctx sdk.Context, gasLimit int64) sdk.Context {
	// Height 0 runs unmetered: consumption is uncapped, so a genesis tx whose
	// gas_wanted is too low still runs to completion instead of being dropped.
	// The fee is charged either way -- genesis txs carry a real gas_fee, and
	// Phase 2b does not exempt them.
	//
	// Simulation is NOT exempt, despite being the case people expect to be.
	// This function is not told whether it is simulating -- there is one call
	// site and it is unconditional -- so a simulated tx above height 0 runs
	// under a real basicGasMeter(GasWanted) like any other. Anything reasoning
	// about what `.app/simulate` can afford has to account for that.
	if ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(store.NewInfiniteGasMeter())
	}

	// Historical tx replay in source-gas mode: bypass the new VM's gas meter
	// so source-chain outcomes are preserved regardless of gas-metering changes.
	if skip, _ := ctx.Value(SkipGasMeteringKey{}).(bool); skip {
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
