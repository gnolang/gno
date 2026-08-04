package bank

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// bank.Keeper defines a module interface that facilitates the transfer of
// coins between accounts without the possibility of creating coins.
type BankKeeperI interface {
	ViewKeeperI

	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) error
	SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error

	SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	SetCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error

	InitGenesis(ctx sdk.Context, data GenesisState)
	GetParams(ctx sdk.Context) Params

	MintCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	BurnCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error
	RecomputeSupply(ctx sdk.Context)
}

var _ BankKeeperI = &BankKeeper{}

// BankKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the BankKeeperI interface.
type BankKeeper struct {
	ViewKeeper

	acck auth.AccountKeeper
	// The keeper used to store parameters
	prmk params.ParamsKeeperI
}

// NewBankKeeper returns a new BankKeeper.
//
// accountDenoms is the allowlist of denoms held inside the account object; see
// balance.go. Everything else gets its own key. Pass the chain's gas denoms.
func NewBankKeeper(acck auth.AccountKeeper, pk params.ParamsKeeperI, key store.StoreKey, accountDenoms []string) BankKeeper {
	return BankKeeper{
		ViewKeeper: NewViewKeeper(acck, key, accountDenoms),
		acck:       acck,
		prmk:       pk,
	}
}

// This is a convenience function for manually setting the restricted denoms.
// Useful for testing and initchain setup.
func (bank BankKeeper) SetRestrictedDenoms(ctx sdk.Context, restrictedDenoms []string) {
	bank.prmk.SetStrings(ctx, "p:restricted_denoms", restrictedDenoms)
}

func (bank BankKeeper) RestrictedDenoms(ctx sdk.Context) []string {
	params := bank.GetParams(ctx)
	return params.RestrictedDenoms
}

type stringSet map[string]struct{}

func toSet(str []string) stringSet {
	ss := stringSet{}

	for _, key := range str {
		ss[key] = struct{}{}
	}
	return ss
}

// InputOutputCoins handles a list of inputs and outputs
func (bank BankKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) error {
	// Safety check ensuring that when sending coins the bank must maintain the
	// Check supply invariant and validity of Coins.
	if err := ValidateInputsOutputs(inputs, outputs); err != nil {
		return err
	}

	for _, in := range inputs {
		if !bank.canSendCoins(ctx, in.Address, in.Coins) {
			return std.RestrictedTransferError{}
		}
		// Per-input session spend check: each Input.Address is a tx
		// signer (per MsgMultiSend.GetSigners), so if any input belongs
		// to a session's master, the input amount counts against that
		// session's SpendLimit. No-op for non-session signers.
		if err := auth.CheckAndDeductSessionSpend(ctx, bank.acck, in.Address, in.Coins); err != nil {
			return err
		}
		if err := bank.SubtractCoins(ctx, in.Address, in.Coins); err != nil {
			return err
		}

		/*
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute(types.AttributeKeySender, in.Address.String()),
				),
			)
		*/
	}

	for _, out := range outputs {
		if err := bank.AddCoins(ctx, out.Address, out.Coins); err != nil {
			return err
		}

		/*
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeTransfer,
					sdk.NewAttribute(types.AttributeKeyRecipient, out.Address.String()),
				),
			)
		*/
	}

	return nil
}

// canSendCoins returns true if the coins can be sent without violating any restriction.
func (bank BankKeeper) canSendCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool {
	rds := bank.RestrictedDenoms(ctx)
	if len(rds) == 0 {
		// No restrictions.
		return true
	}
	if amt.ContainOneOfDenom(toSet(rds)) {
		acc := bank.acck.GetAccount(ctx, addr)
		accr, ok := acc.(std.AccountUnrestricter)
		if ok && accr.IsTokenLockWhitelisted() {
			return true
		}
		return false
	}
	return true
}

// SendCoins moves coins from one account to another, restrction could be applied
func (bank BankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	// If amt is zero do nothing.
	if amt.IsZero() {
		return nil
	}

	// read restricted boolean value from param.IsRestrictedTransfer()
	// canSendCoins is true until they have agreed to the waiver
	if !bank.canSendCoins(ctx, fromAddr, amt) {
		return std.RestrictedTransferError{}
	}

	// If the tx is session-signed and fromAddr is the session's master,
	// deduct from the session's SpendLimit. No-op otherwise.
	// SendCoinsUnrestricted deliberately bypasses this (gas collection,
	// storage deposit refunds).
	if err := auth.CheckAndDeductSessionSpend(ctx, bank.acck, fromAddr, amt); err != nil {
		return err
	}

	return bank.sendCoins(ctx, fromAddr, toAddr, amt)
}

// SendCoinsUnrestricted is used for paying gas.
// It bypasses vesting and session-spend checks.
func (bank BankKeeper) SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	if err := bank.subtractCoinsUnrestricted(ctx, fromAddr, amt); err != nil {
		return err
	}
	return bank.AddCoins(ctx, toAddr, amt)
}

func (bank BankKeeper) sendCoins(
	ctx sdk.Context,
	fromAddr crypto.Address,
	toAddr crypto.Address,
	amt std.Coins,
) error {
	if err := bank.SubtractCoins(ctx, fromAddr, amt); err != nil {
		return err
	}

	if err := bank.AddCoins(ctx, toAddr, amt); err != nil {
		return err
	}

	/*
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, toAddr.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
			),
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
			),
		})
	*/

	return nil
}

// upgradeVestingAccount replaces a fully-vested VestingAccount with a plain
// BaseAccount. Returns the replacement account, or the original if not ready.
func (bank BankKeeper) upgradeVestingAccount(ctx sdk.Context, acc std.Account) (std.Account, bool) {
	va, ok := acc.(std.VestingAccount)
	if !ok {
		return acc, false
	}
	if !va.GetVestingCoins(ctx.BlockTime()).IsZero() {
		return acc, false
	}
	// Deliberately does not persist. Callers write it on their success path only:
	// this used to SetAccount here, which meant a debit that then failed its
	// affordability check still rewrote the account object — a write on an operation
	// that reported failure. The caller knows whether it succeeded; this does not.
	return &std.BaseAccount{
		Address:       va.GetAddress(),
		Coins:         va.GetCoins(),
		PubKey:        va.GetPubKey(),
		AccountNumber: va.GetAccountNumber(),
		Sequence:      va.GetSequence(),
	}, true
}

// splitByTier partitions amt by where each denom is stored. Both halves keep
// amt's ascending order, so each can be used directly as std.Coins.
func (view ViewKeeper) splitByTier(amt std.Coins) (split, account std.Coins) {
	for _, coin := range amt {
		if view.inAccountTier(coin.Denom) {
			account = append(account, coin)
		} else {
			split = append(split, coin)
		}
	}
	return split, account
}

// ensureAccount creates addr's account if it has none, and returns it.
//
// Receiving coins has always created the account, and that must not change: an
// address with no account cannot sign a transaction (auth.GetSignerAcc), so a
// recipient funded without one would hold visible but permanently unspendable
// coins. Account creation also allocates an account number from a global
// counter, so *when* it happens is consensus state.
//
// A failed credit does not consume a number, and that holds by ordering rather than by
// a guard: this runs before AddCoins' only fallible step, but both reachable failures —
// a stored account that refuses coins, and an overflowing balance — require an account
// to exist already, so nothing is created and then abandoned. Measured. The property
// would break if a freshly created account could refuse coins, which BaseAccount cannot
// (its SetCoins returns nil unconditionally); no test can pin that, because the state
// is unconstructible, so the ordering here is the guard.
func (bank BankKeeper) ensureAccount(ctx sdk.Context, addr crypto.Address) std.Account {
	acc := bank.acck.GetAccount(ctx, addr)
	if acc == nil {
		acc = bank.acck.NewAccountWithAddress(ctx, addr)
		bank.acck.SetAccount(ctx, acc)
	}
	return acc
}

// setAccountTierCoins replaces the account-tier balances in addr's account
// object. acc is the account if the caller already has it, else nil.
func (bank BankKeeper) setAccountTierCoins(ctx sdk.Context, acc std.Account, addr crypto.Address, coins std.Coins) error {
	if acc == nil {
		acc = bank.ensureAccount(ctx, addr)
	}
	if err := acc.SetCoins(coins); err != nil {
		return err
	}
	bank.acck.SetAccount(ctx, acc)
	return nil
}

// SubtractCoins subtracts amt from the coins at the addr.
//
// Enforces vesting: if the account is a VestingAccount, the amount must not
// exceed the spendable (unlocked) balance at the current block time.
func (bank BankKeeper) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}

	// Once the schedule completes the account becomes a plain BaseAccount, so
	// later transfers skip the vesting check entirely.
	acc, upgraded := bank.upgradeVestingAccount(ctx, bank.acck.GetAccount(ctx, addr))

	// The schedule is checked one denom at a time, reading only the denoms being
	// spent — never the whole balance set, which is what makes this O(len(amt))
	// rather than O(n). It is deliberately tier-agnostic: a genesis balances file
	// may name a "/"-prefixed denom (ValidateDenom accepts one), so locking is
	// not assumed to apply only to the account-object tier. GetCoin routes
	// each denom to wherever it actually lives.
	if va, ok := acc.(std.VestingAccount); ok {
		locked := va.LockedCoins(ctx.BlockTime())
		for _, coin := range amt {
			lockedAmt := locked.AmountOf(coin.Denom)
			if lockedAmt == 0 {
				continue
			}
			// The schedule can lock more than the account holds, once an
			// unrestricted transfer has spent into the locked portion. Clamp so
			// the error reports nothing spendable rather than a negative.
			spendable := max(bank.GetCoin(ctx, addr, coin.Denom)-lockedAmt, 0)
			if spendable < coin.Amount {
				return std.ErrVestingLockedCoins(fmt.Sprintf(
					"insufficient spendable coins; %d%s < %s (locked=%s)",
					spendable, coin.Denom, coin, locked,
				))
			}
		}
	}

	return bank.subtract(ctx, acc, addr, amt, upgraded)
}

// subtractCoinsUnrestricted performs raw coin subtraction without vesting or
// session-spend enforcement. Used for gas payments, storage deposit refunds,
// and other system-level transfers.
//
// Still upgrades fully-vested accounts to BaseAccount when the schedule ends.
func (bank BankKeeper) subtractCoinsUnrestricted(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}
	acc, upgraded := bank.upgradeVestingAccount(ctx, bank.acck.GetAccount(ctx, addr))
	return bank.subtract(ctx, acc, addr, amt, upgraded)
}

// subtract debits amt: one key per split-tier denom, plus at most one account
// write for the account-tier denoms. No vesting or session checks.
//
// acc is the account as the caller already read it, or nil. Threaded in only to
// avoid re-decoding it; pass nil if you do not have it.
//
// Relies on amt holding each denom at most once, which std.Coins guarantees and
// every caller checks via amt.IsValid(). A repeated denom would have both
// entries compute their remainder from the same starting balance, so the last
// write would win and one debit would vanish.
//
// Every debit is checked before any is written, so returning an error leaves
// balances untouched. The previous whole-account implementation validated the
// resulting Coins before its single write and was therefore atomic by
// construction; nothing here should be less safe. A transaction abort would
// usually roll a partial debit back, but that puts the invariant in the caller,
// and it does not hold for a caller that logs and continues.
func (bank BankKeeper) subtract(ctx sdk.Context, acc std.Account, addr crypto.Address, amt std.Coins, upgraded bool) error {
	// A non-positive amount turns this debit into a credit, on either tier, so
	// reject it before splitting, along with a repeated denom: two entries would
	// each be computed from the same starting balance, so only one debit would
	// land. Every caller validates amt, so this cannot fire today; it is here
	// because neither tier defends itself otherwise:
	//
	//   - split tier: the check is `old < coin.Amount`, which a negative passes
	//     since old is never negative. The subtraction then credits, or overflows
	//     to a negative that only encodeBalance catches.
	//   - account tier: Coins.SubUnsafe followed by IsValid validates the
	//     *result*, and a negative debit produces a larger positive result, which
	//     is perfectly valid. Subtracting -500 from 100 yields 600.
	//
	// With every amount positive, and each checked against the balance below, no
	// subtraction here can overflow.
	if err := amt.Validate(); err != nil {
		return err
	}
	// acc must be addr's account: it is threaded in to save a read, and nothing in
	// the types connects the two, so a mismatched pair would debit the account tier
	// of one address and the split tier of another.
	if acc != nil && acc.GetAddress() != addr {
		return fmt.Errorf("account %s was passed for address %s", acc.GetAddress(), addr)
	}

	split, account := bank.splitByTier(amt)

	// Carry each new balance with its denom rather than in a positionally
	// aligned slice: an index mistake here would write one denom's balance to
	// another denom's key, which no type would catch.
	debited := make(std.Coins, len(split))
	for i, coin := range split {
		old := bank.getSplitBalance(ctx, addr, coin.Denom)
		if old < coin.Amount {
			return std.ErrInsufficientCoins(fmt.Sprintf(
				"insufficient account funds; %d%s < %s", old, coin.Denom, coin))
		}
		debited[i] = std.Coin{Denom: coin.Denom, Amount: old - coin.Amount}
	}

	var newCoins std.Coins
	if len(account) > 0 {
		oldCoins := std.NewCoins()
		if acc != nil {
			oldCoins = bank.accountTierCoins(acc)
		}
		newCoins = oldCoins.SubUnsafe(account)
		if !newCoins.IsValid() {
			return std.ErrInsufficientCoins(fmt.Sprintf(
				"insufficient account funds; %s < %s", oldCoins, account))
		}
	}

	// A completed vesting schedule is collapsed to a plain account here, on the
	// success path, rather than when it was detected — so a failed debit leaves the
	// account object untouched. setAccountTierCoins below may write it again; the
	// cache store refund-dedups repeated writes, so the cost is unchanged.
	if upgraded {
		bank.acck.SetAccount(ctx, acc)
	}

	// The account write goes first because it is the only step that can still
	// fail — setAccountTierCoins returns Account.SetCoins's error — while
	// setSplitBalance cannot, every balance having been checked non-negative
	// above. With the fallible step first, a failure leaves nothing written.
	// AddCoins orders its writes the same way, for the same reason.
	if len(account) > 0 {
		if err := bank.setAccountTierCoins(ctx, acc, addr, newCoins); err != nil {
			return err
		}
	}
	for _, coin := range debited {
		bank.setSplitBalance(ctx, addr, coin.Denom, coin.Amount)
	}
	return nil
}

// AddCoins adds amt to the coins at the addr.
func (bank BankKeeper) AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}
	split, account := bank.splitByTier(amt)

	// Sum every new balance before writing any, for the same reason subtract
	// checks before writing: a failure partway must not leave some credits
	// applied and others not.
	credited := make(std.Coins, len(split))
	for i, coin := range split {
		old := bank.getSplitBalance(ctx, addr, coin.Denom)
		sum, ok := overflow.Add(old, coin.Amount)
		if !ok {
			// Matches std.Coins.Add, which panics rather than wrapping. A silent
			// wrap here would mint coins out of a negative balance.
			panic(fmt.Sprintf("coin add overflow: %d%s + %s", old, coin.Denom, coin))
		}
		credited[i] = std.Coin{Denom: coin.Denom, Amount: sum}
	}

	// Receiving coins creates the account whichever tier they land in.
	acc := bank.ensureAccount(ctx, addr)

	// The account-tier credit goes first: Coins.Add panics on overflow, so doing
	// it before the split writes keeps the whole credit all-or-nothing.
	if len(account) > 0 {
		if err := acc.SetCoins(bank.accountTierCoins(acc).Add(account)); err != nil {
			return err
		}
		bank.acck.SetAccount(ctx, acc)
	}
	for _, coin := range credited {
		bank.setSplitBalance(ctx, addr, coin.Denom, coin.Amount)
	}
	return nil
}

// SetCoins replaces every balance at addr.
//
// Replace-all is only meaningful at genesis and in tests; ordinary transfers use
// AddCoins/SubtractCoins, which touch one key per split-tier denom moved. Balances
// present but absent from amt are deleted, so this is a replacement and not a
// merge.
//
// Does not maintain the supply counter. It cannot: InitChainer writes the account
// object with the full pre-split amount before calling this, so the old value read
// here equals the new one and any delta would be zero for a vesting account. Call
// RecomputeSupply after a batch of these; see supply.go.
//
// Stays exported despite that, because genesis lives in another package
// (gno.land/pkg/gnoland) and has to call it. The protection that matters is already
// structural rather than lexical: SetCoins is absent from both vm.BankKeeperI and
// auth.BankKeeperI, so neither a realm nor the ante handler can reach it. Only a new
// tm2-side handler holding BankKeeperI could, which is what the warning below is for.
//
// Cost is O(denoms currently held), and each removal is a full store write, so
// this is not safe to call on an address whose denom count an attacker controls —
// clearing a few hundred would exceed the block gas limit. Every caller today
// runs at genesis, where the count is known.
func (bank BankKeeper) SetCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	if !amt.IsValid() {
		return std.ErrInvalidCoins(amt.String())
	}
	split, account := bank.splitByTier(amt)

	// The account tier goes first, as in subtract and AddCoins: it is the only step
	// that can fail, and with it last a failure would return an error having already
	// replaced the split tier.
	if err := bank.setAccountTierCoins(ctx, nil, addr, account); err != nil {
		return err
	}
	// splitCoins fully drains and closes its iterator before returning, so there
	// is no live iteration to invalidate by deleting here.
	for _, coin := range bank.splitCoins(ctx, addr) {
		bank.setSplitBalance(ctx, addr, coin.Denom, 0)
	}
	for _, coin := range split {
		bank.setSplitBalance(ctx, addr, coin.Denom, coin.Amount)
	}
	return nil
}

// ----------------------------------------
// ViewKeeper

// ViewKeeperI defines a module interface that facilitates read only access to
// account balances.
type ViewKeeperI interface {
	GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins
	GetCoin(ctx sdk.Context, addr crypto.Address, denom string) int64
	HasCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool
	TotalSupply(ctx sdk.Context, denom string) int64
}

var _ ViewKeeperI = ViewKeeper{}

// ViewKeeper implements a read only keeper implementation of ViewKeeperI.
type ViewKeeper struct {
	acck auth.AccountKeeper
	// key holds split-tier balances; see balance.go.
	key store.StoreKey
	// accountDenoms is the allowlist of denoms kept in the account object.
	accountDenoms map[string]struct{}
}

// NewViewKeeper returns a new ViewKeeper.
//
// accountDenoms is the allowlist of denoms held inside the account object.
func NewViewKeeper(acck auth.AccountKeeper, key store.StoreKey, accountDenoms []string) ViewKeeper {
	allow := make(map[string]struct{}, len(accountDenoms))
	for _, denom := range accountDenoms {
		// An entry here is a security decision, not a config detail. The account
		// object is read and rewritten on every transaction its owner sends (the
		// ante handler bumps the sequence) and its value is metered per byte, so
		// every denom in this tier is paid for on every transaction that address
		// ever makes. What matters is therefore who can put a denom into a
		// stranger's blob: sending ugnot costs the sender the coins, an IBC
		// voucher costs a real transfer, but IssueCoin mints a realm denom from
		// nothing to any address with no recipient check. Allowlisting one would
		// let its issuer add ~40 bytes — roughly 1,200 gas per transaction at
		// 17/byte read plus 14/byte write — to any account it picks, for free.
		// The victim can clear it by transferring the balance away, since a zero
		// drops out of the Coins set, but that costs them a transaction and the
		// issuer can re-mint for less: a griefing loop, not a permanent brick.
		// Exact matching holds it to one denom rather than unlimited ones, which
		// is what keeps it a tax. Reject at construction, where the chain fails to
		// start, rather than at the first mint.
		if err := std.ValidateDenom(denom); err != nil {
			panic(fmt.Sprintf("invalid account-tier denom %q: %v", denom, err))
		}
		if std.IsRealmDenom(denom) {
			panic(fmt.Sprintf("account-tier denom %q is realm-issuable", denom))
		}
		allow[denom] = struct{}{}
	}
	return ViewKeeper{acck: acck, key: key, accountDenoms: allow}
}

// inAccountTier reports whether denom's balance lives inside the account object
// rather than in its own key.
//
// An allowlist, not a pattern: the account tier exists only because the account
// object is written on every transaction anyway (for the sequence bump), so a
// balance kept there rides along at no extra cost. That argument holds for the
// denoms used to pay gas and for nothing else. Anything absent from the list —
// realm-issued coins, IBC vouchers, any future denom — gets its own key, where
// its count cannot affect the holder's other transactions.
//
// Deliberately closed rather than open: a new kind of denom is split by default,
// so arriving at the account tier requires an explicit decision.
func (view ViewKeeper) inAccountTier(denom string) bool {
	_, ok := view.accountDenoms[denom]
	return ok
}

// accountTierCoins returns the balances held inside acc, asserting that every
// one of them belongs to the account tier. Returns nil for a nil account.
//
// This is the mirror of the assertion in splitCoins, and it covers the direction
// the allowlist is actually expected to move: shrinking it is how a chain
// migrates to a fully split layout. Run such a binary against unmigrated state
// and a denom sits in the account object while the keeper believes it is split —
// GetCoins would sum both homes while GetCoin saw only one, so a balance would
// read as spendable that is not. A denom must live in exactly one tier; fail
// loudly at the first read rather than report a number nobody can spend.
func (view ViewKeeper) accountTierCoins(acc std.Account) std.Coins {
	if acc == nil {
		return nil
	}
	coins := acc.GetCoins()
	for _, coin := range coins {
		if !view.inAccountTier(coin.Denom) {
			panic(fmt.Sprintf(
				"denom %q is in the account object but not in the account tier: "+
					"the allowlist changed without migrating existing balances", coin.Denom))
		}
	}
	return coins
}

// Logger returns a module-specific logger.
func (view ViewKeeper) Logger(ctx sdk.Context) *slog.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}

// GetCoins returns every coin held at addr, from both tiers.
//
// The two are merged rather than concatenated. An earlier version could
// concatenate, because the split tier held only "/"-prefixed denoms which sort
// below everything else; with an allowlist that no longer holds — a split denom
// such as "atom" sorts before an account-tier "ugnot" — so neither order is
// universally ascending. Coins.Add is a merge over sorted sets, and the tiers
// are disjoint by construction, so no two amounts are ever summed.
//
// Costs O(number of split-tier denoms held). Use GetCoin when one denom will
// do — that is the whole point of the split.
func (view ViewKeeper) GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins {
	split := view.splitCoins(ctx, addr)
	account := view.accountTierCoins(view.acck.GetAccount(ctx, addr))
	switch {
	case len(split) == 0 && len(account) == 0:
		return std.NewCoins()
	case len(split) == 0:
		// The common case on a chain whose only account-tier denom is the gas
		// denom: no merge, no copy.
		return account
	case len(account) == 0:
		return split
	}
	// AddUnsafe, not Add. Add revalidates every denom in the result, and there is
	// nothing for that to find: each denom passed ValidateDenom on the way in and
	// the tiers are disjoint, so the only property iteration could break is
	// ordering, which splitCoins asserts itself. Add would also panic rather than
	// return, and bank/balances reaches this unauthenticated — corrupt state
	// belongs in the answer, where the invariants can report it, not in a panic.
	return split.AddUnsafe(account)
}

// GetCoin returns addr's balance of one denom without reading any other. This
// is the O(1) accessor.
func (view ViewKeeper) GetCoin(ctx sdk.Context, addr crypto.Address, denom string) int64 {
	// Bound the denom before touching anything. No balance can exist for an
	// over-long denom, since every write validates, so returning zero is exact —
	// and this is the only thing standing between a caller-supplied string and
	// work proportional to its length: inAccountTier hashes the whole string,
	// BalanceKey copies it, and the cache store retains a copy of the key for the
	// life of the transaction while charging a flat rate. The bound has to be
	// checked first for any of that to be avoided.
	if len(denom) > std.MaxDenomLength {
		return 0
	}
	if !view.inAccountTier(denom) {
		return view.getSplitBalance(ctx, addr, denom)
	}
	return view.accountTierCoins(view.acck.GetAccount(ctx, addr)).AmountOf(denom)
}

// HasCoins returns whether or not an account has at least amt coins.
//
// Checked per denom so that holding many unrelated denoms costs nothing here. An
// empty amt is satisfied, as under the previous IsAllGTE. The two differ on one
// unreachable edge — a zero-amount entry against an empty balance set, which
// IsAllGTE rejected and this accepts — because valid std.Coins never carry a
// zero amount.
func (view ViewKeeper) HasCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool {
	for _, coin := range amt {
		if view.GetCoin(ctx, addr, coin.Denom) < coin.Amount {
			return false
		}
	}
	return true
}
