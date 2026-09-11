package std

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// -----------------------------------------------------------------------------
// Coin

// Coin hold some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// NewCoin returns a new coin with a denomination and amount.
// It will panic if the amount is negative.
// To construct a negative (invalid) amount, use an operation.
func NewCoin(denom string, amount int64) Coin {
	coin, err := NewCoinSafe(denom, amount)
	if err != nil {
		panic(err)
	}

	return coin
}

// NewCoinSafe returns a new coin with a denomination and amount.
// It will return an error if the amount is negative.
// To construct a negative (invalid) amount, use an operation.
func NewCoinSafe(denom string, amount int64) (Coin, error) {
	if err := validate(denom, amount); err != nil {
		return Coin{}, err
	}

	return Coin{
		Denom:  denom,
		Amount: amount,
	}, nil
}

func (coin Coin) MarshalAmino() (string, error) {
	return coin.String(), nil
}

func (coin *Coin) UnmarshalAmino(coinstr string) (err error) {
	if coinstr == "" {
		return nil
	}
	coin2, err := ParseCoin(coinstr)
	if err != nil {
		return err
	}
	*coin = coin2
	return nil
}

// String provides a human-readable representation of a coin
func (coin Coin) String() string {
	if coin.IsZero() {
		return ""
	} else {
		return fmt.Sprintf("%d%v", coin.Amount, coin.Denom)
	}
}

// validate returns an error if the Coin has a negative amount or if
// the denom is invalid.
func validate(denom string, amount int64) error {
	if err := ValidateDenom(denom); err != nil {
		return err
	}

	if amount < 0 {
		return fmt.Errorf("negative coin amount: %d", amount)
	}

	return nil
}

// IsValid returns true if the Coin has a non-negative amount and the denom is valid.
func (coin Coin) IsValid() bool {
	return validate(coin.Denom, coin.Amount) == nil
}

// IsZero returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount == 0
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (coin Coin) IsGTE(other Coin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}
	return coin.Amount >= other.Amount
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (coin Coin) IsLT(other Coin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}
	return coin.Amount < other.Amount
}

// IsEqual returns true if the two sets of Coins have the same value
func (coin Coin) IsEqual(other Coin) bool {
	if coin.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, other.Denom))
	}
	return coin.Amount == other.Amount
}

// Adds amounts of two coins with same denom.
// If the coins differ in denom then it panics.
// An overflow or underflow panics.
// An invalid result panics.
func (coin Coin) Add(coinB Coin) Coin {
	res := coin.AddUnsafe(coinB)
	if err := validate(res.Denom, res.Amount); err != nil {
		panic(fmt.Sprintf("invalid result: %v + %v = %v: %v", coin, coinB, res, err))
	}
	return res
}

func (coin Coin) AddUnsafe(coinB Coin) Coin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}
	sum, ok := overflow.Add(coin.Amount, coinB.Amount)
	if !ok {
		panic(fmt.Sprintf("coin add overflow/underflow: %v, %v", coin, coinB))
	}
	return Coin{coin.Denom, sum}
}

// Subtracts amounts of two coins with same denom.
// If the coins differ in denom then it panics.
// An overflow or underflow panics.
// An invalid result panics.
func (coin Coin) Sub(coinB Coin) Coin {
	res := coin.SubUnsafe(coinB)
	if err := validate(res.Denom, res.Amount); err != nil {
		panic(fmt.Sprintf("invalid result: %v - %v = %v: %v", coin, coinB, res, err))
	}
	return res
}

func (coin Coin) SubUnsafe(coinB Coin) Coin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}
	dff, ok := overflow.Sub(coin.Amount, coinB.Amount)
	if !ok {
		panic(fmt.Sprintf("coin subtract overflow/underflow: %v, %v", coin, coinB))
	}
	return Coin{coin.Denom, dff}
}

// IsPositive returns true if coin amount is positive.
func (coin Coin) IsPositive() bool {
	return coin.Amount > 0
}

// IsNegative returns true if the coin amount is negative and false otherwise.
func (coin Coin) IsNegative() bool {
	return coin.Amount < 0
}

// -----------------------------------------------------------------------------
// Coins

// Coins is a set of Coin, one per currency
type Coins []Coin

// NewCoins constructs a new coin set.
func NewCoins(coins ...Coin) Coins {
	// Sort below reorders in place, and with `NewCoins(myVar...)` the variadic
	// array is the caller's own slice, so sorting it would silently reorder what
	// the caller still holds. Copy first. The Gno mirror in chain/coins.gno
	// copies for the same reason.
	cz := make(Coins, len(coins))
	copy(cz, coins)

	// remove zeroes
	newCoins := removeZeroCoins(cz)
	if len(newCoins) == 0 {
		return Coins{}
	}

	newCoins.Sort()

	// detect duplicate Denoms
	if dupIndex := findDup(newCoins); dupIndex != -1 {
		panic(fmt.Errorf("find duplicate denom: %s", newCoins[dupIndex]))
	}

	if err := newCoins.validate(); err != nil {
		panic(fmt.Errorf("invalid coin set %s: %w", newCoins, err))
	}

	return newCoins
}

func (coins Coins) MarshalAmino() (string, error) {
	return coins.String(), nil
}

func (coins *Coins) UnmarshalAmino(coinsstr string) (err error) {
	coins2, err := ParseCoins(coinsstr)
	if err != nil {
		return err
	}
	*coins = coins2
	return nil
}

func (coins Coins) String() string {
	if len(coins) == 0 {
		return ""
	}

	var out strings.Builder
	for _, coin := range coins {
		out.WriteString(fmt.Sprintf("%v,", coin.String()))
	}
	return out.String()[:len(out.String())-1]
}

// IsValid asserts the Coins are sorted, have positive amount,
// and Denom does not contain upper case characters.
func (coins Coins) IsValid() bool {
	return coins.validate() == nil
}

// Validate is IsValid with the reason. Callers rejecting a caller-supplied set
// want to say why.
func (coins Coins) Validate() error {
	return coins.validate()
}

// validate checks that the Coins are sorted, have positive amounts,
// and valid denoms. Returns an error describing the issue if invalid.
func (coins Coins) validate() error {
	switch len(coins) {
	case 0:
		return nil
	case 1:
		if err := ValidateDenom(coins[0].Denom); err != nil {
			return err
		}
		if !coins[0].IsPositive() {
			return fmt.Errorf("non-positive coin amount: %d", coins[0].Amount)
		}
		return nil
	default:
		// check single coin case
		if err := (Coins{coins[0]}).validate(); err != nil {
			return err
		}

		lowDenom := coins[0].Denom
		for _, coin := range coins[1:] {
			if err := ValidateDenom(coin.Denom); err != nil {
				return err
			}
			if coin.Denom < lowDenom {
				return fmt.Errorf("coins not sorted: %s < %s", coin.Denom, lowDenom)
			}
			if coin.Denom == lowDenom {
				return fmt.Errorf("duplicate denom: %s", coin.Denom)
			}
			if !coin.IsPositive() {
				return fmt.Errorf("non-positive coin amount: %d", coin.Amount)
			}

			// we compare each coin against the last denom
			lowDenom = coin.Denom
		}

		return nil
	}
}

// Add adds two sets of coins.
//
// e.g.
// {2A} + {A, 2B} = {3A, 2B}
// {2A} + {0B} = {2A}
//
// NOTE: Add operates under the invariant that coins are sorted by
// denominations. Panics on invalid result.
func (coins Coins) Add(coinsB Coins) Coins {
	res := coins.AddUnsafe(coinsB)
	if err := res.validate(); err != nil {
		panic(fmt.Sprintf("invalid result: %v + %v = %v: %v", coins, coinsB, res, err))
	}
	return res
}

// AddUnsafe will perform addition of two coins sets. If both coin sets are
// empty, then an empty set is returned. If only a single set is empty, the
// other set is returned. Otherwise, the coins are compared in order of their
// denomination and addition only occurs when the denominations match, otherwise
// the coin is simply added to the sum assuming it's not zero.
func (coins Coins) AddUnsafe(coinsB Coins) Coins {
	sum := ([]Coin)(nil)
	indexA, indexB := 0, 0
	lenA, lenB := len(coins), len(coinsB)

	for {
		if indexA == lenA {
			if indexB == lenB {
				// return nil coins if both sets are empty
				return sum
			}

			// return set B (excluding zero coins) if set A is empty
			return append(sum, removeZeroCoins(coinsB[indexB:])...)
		} else if indexB == lenB {
			// return set A (excluding zero coins) if set B is empty
			return append(sum, removeZeroCoins(coins[indexA:])...)
		}

		coinA, coinB := coins[indexA], coinsB[indexB]

		switch strings.Compare(coinA.Denom, coinB.Denom) {
		case -1: // coin A denom < coin B denom
			if !coinA.IsZero() {
				sum = append(sum, coinA)
			}

			indexA++

		case 0: // coin A denom == coin B denom
			res := coinA.AddUnsafe(coinB)
			if !res.IsZero() {
				sum = append(sum, res)
			}

			indexA++
			indexB++

		case 1: // coin A denom > coin B denom
			if !coinB.IsZero() {
				sum = append(sum, coinB)
			}

			indexB++
		}
	}
}

// ContainOneOfDenom check if a Coins instance contains a denom in the provided denomos
func (coins Coins) ContainOneOfDenom(denoms map[string]struct{}) bool {
	if len(denoms) == 0 {
		return false
	}

	for _, coin := range coins {
		if _, ok := denoms[coin.Denom]; ok && coin.IsPositive() {
			return true
		}
	}

	return false
}

// DenomsSubsetOf returns true if receiver's denom set
// is subset of coinsB's denoms.
func (coins Coins) DenomsSubsetOf(coinsB Coins) bool {
	// more denoms in B than in receiver
	if len(coins) > len(coinsB) {
		return false
	}

	for _, coin := range coins {
		if coinsB.AmountOf(coin.Denom) == 0 {
			return false
		}
	}

	return true
}

// Sub subtracts a set of coins from another.
//
// e.g.
// {2A, 3B} - {A} = {A, 3B}
// {2A} - {0B} = {2A}
// {A, B} - {A} = {B}
//
// Panics on invalid result.
func (coins Coins) Sub(coinsB Coins) Coins {
	res := coins.SubUnsafe(coinsB)
	if err := res.validate(); err != nil {
		panic(fmt.Sprintf("invalid result: %v - %v = %v: %v", coins, coinsB, res, err))
	}
	return res
}

// SubUnsafe performs the same arithmetic as Sub but returns a boolean if any
// negative coin amount was returned.
func (coins Coins) SubUnsafe(coinsB Coins) Coins {
	res := coins.AddUnsafe(coinsB.negative())
	return res
}

// IsAllGT returns true if for every denom in coinsB,
// the denom is present at a greater amount in coins.
func (coins Coins) IsAllGT(coinsB Coins) bool {
	if len(coins) == 0 {
		return false
	}

	if len(coinsB) == 0 {
		return true
	}

	if !coinsB.DenomsSubsetOf(coins) {
		return false
	}

	for _, coinB := range coinsB {
		amountA, amountB := coins.AmountOf(coinB.Denom), coinB.Amount
		if amountA <= amountB {
			return false
		}
	}

	return true
}

// IsAllGTE returns false if for any denom in coinsB,
// the denom is present at a smaller amount in coins;
// else returns true.
func (coins Coins) IsAllGTE(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return true
	}

	if len(coins) == 0 {
		return false
	}

	for _, coinB := range coinsB {
		if coinB.Amount > coins.AmountOf(coinB.Denom) {
			return false
		}
	}

	return true
}

// IsAllLT returns True iff for every denom in coins, the denom is present at
// a smaller amount in coinsB.
func (coins Coins) IsAllLT(coinsB Coins) bool {
	return coinsB.IsAllGT(coins)
}

// IsAllLTE returns true iff for every denom in coins, the denom is present at
// a smaller or equal amount in coinsB.
func (coins Coins) IsAllLTE(coinsB Coins) bool {
	return coinsB.IsAllGTE(coins)
}

// IsAnyGT returns true iff for any denom in coins, the denom is present at a
// greater amount in coinsB.
//
// e.g.
// {2A, 3B}.IsAnyGT{A} = true
// {2A, 3B}.IsAnyGT{5C} = false
// {}.IsAnyGT{5C} = false
// {2A, 3B}.IsAnyGT{} = false
func (coins Coins) IsAnyGT(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return false
	}

	for _, coin := range coins {
		amt := coinsB.AmountOf(coin.Denom)
		if coin.Amount > amt && amt != 0 {
			return true
		}
	}

	return false
}

// IsAnyGTE returns true iff coins contains at least one denom that is present
// at a greater or equal amount in coinsB; it returns false otherwise.
//
// NOTE: IsAnyGTE operates under the invariant that both coin sets are sorted
// by denominations and there exists no zero coins.
func (coins Coins) IsAnyGTE(coinsB Coins) bool {
	if len(coinsB) == 0 {
		return false
	}

	for _, coin := range coins {
		amt := coinsB.AmountOf(coin.Denom)
		if coin.Amount >= amt && amt != 0 {
			return true
		}
	}

	return false
}

// IsZero returns true if there are no coins or all coins are zero.
func (coins Coins) IsZero() bool {
	for _, coin := range coins {
		if !coin.IsZero() {
			return false
		}
	}
	return true
}

// IsEqual returns true if the two sets of Coins have the same value
func (coins Coins) IsEqual(coinsB Coins) bool {
	if len(coins) != len(coinsB) {
		return false
	}

	// XXX instead, consider IsValid() on both and panic if not.
	coins = coins.Sort()
	coinsB = coinsB.Sort()

	for i := range coins {
		if !coins[i].IsEqual(coinsB[i]) {
			return false
		}
	}

	return true
}

// Empty returns true if there are no coins and false otherwise.
func (coins Coins) Empty() bool {
	return len(coins) == 0
}

// Returns the amount of a denom from coins, which may be negative.
//
// The set must be sorted. This is a binary search, so on an unsorted set it can
// step past a denom that is there and return 0 without saying so.
func (coins Coins) AmountOf(denom string) int64 {
	mustValidateDenom(denom)

	switch len(coins) {
	case 0:
		return 0

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return coin.Amount
		}
		return 0

	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]

		if denom < coin.Denom {
			return coins[:midIdx].AmountOf(denom)
		} else if denom == coin.Denom {
			return coin.Amount
		} else {
			return coins[midIdx+1:].AmountOf(denom)
		}
	}
}

// IsAllPositive returns true if there is at least one coin and all currencies
// have a positive value.
// NOTE: besides this function, which is zero sensitive, all other functions
// don't need to be called "IsAll*" -- TODO: rename back coins.IsAll* to coins.Is*?
func (coins Coins) IsAllPositive() bool {
	if len(coins) == 0 {
		return false
	}

	for _, coin := range coins {
		if !coin.IsPositive() {
			return false
		}
	}

	return true
}

// IsAnyNegative returns true if there is at least one coin whose amount
// is negative; returns false otherwise. It returns false if the coin set
// is empty too.
//
// TODO: Remove once unsigned integers are used.
func (coins Coins) IsAnyNegative() bool {
	for _, coin := range coins {
		if coin.IsNegative() {
			return true
		}
	}

	return false
}

// negative returns a set of coins with all amount negative.
//
// Panics rather than wrapping, like the arithmetic that uses it. MinInt64 has no
// positive counterpart, so negating it silently returns MinInt64 again; SubUnsafe
// subtracts by adding this, and would then add where it meant to subtract and
// report a plausible wrong number instead of the overflow it really hit.
//
// TODO: Remove once unsigned integers are used.
func (coins Coins) negative() Coins {
	res := make([]Coin, 0, len(coins))

	for _, coin := range coins {
		amount, ok := overflow.Sub(0, coin.Amount)
		if !ok {
			panic(fmt.Sprintf("coin negate overflow: %v", coin))
		}
		res = append(res, Coin{
			Denom:  coin.Denom,
			Amount: amount,
		})
	}

	return res
}

// removeZeroCoins removes all zero coins from the given coin set in-place.
// removeZeroCoins returns coins without its zero-amount entries.
//
// The argument is never modified. Callers pass a view of data they do not own:
// [Coins.AddUnsafe] passes both its receiver and its argument, and [NewCoins]
// passes the variadic array, which is the caller's own slice when spread with
// `NewCoins(myVar...)`. Deleting in place shifts the rest of the set down over
// the gap, so the caller would come back holding a reordered set with a
// zero-amount coin at the end that it never had -- one that no longer passes
// [Coins.Validate].
//
// Returns the input as-is when there is nothing to remove, which is every call
// on a valid set, since validate rejects a zero amount. So the common path
// still does not allocate.
func removeZeroCoins(coins Coins) Coins {
	zeros := 0
	for _, coin := range coins {
		if coin.IsZero() {
			zeros++
		}
	}
	if zeros == 0 {
		return coins
	}

	res := make(Coins, 0, len(coins)-zeros)
	for _, coin := range coins {
		if !coin.IsZero() {
			res = append(res, coin)
		}
	}
	return res
}

// -----------------------------------------------------------------------------
// Sort interface

func (coins Coins) Len() int           { return len(coins) }
func (coins Coins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }
func (coins Coins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = Coins{}

// Sort is a helper function to sort the set of coins inplace
func (coins Coins) Sort() Coins {
	sort.Sort(coins)
	return coins
}

// -----------------------------------------------------------------------------
// Parsing

var (
	// A realm denom embeds a package path verbatim (chain.CoinDenom), so this
	// charset has to stay a superset of rePkgPathURL's charset in memfile.go.
	// "-" was the one character missing from it. Package paths admit "-" in the
	// domain and in namespace segments, so a realm at gno.land/r/my-org/token
	// deployed fine and then failed on its first IssueCoin — a silent, late
	// failure. TestValidateDenomAcceptsDeployablePaths pins that relation for
	// every character class in rePkgPathURL. The trailing "-" is escaped so
	// that appending another character cannot silently turn it into a range.
	//
	// "-" is deliberately absent from the leading class: a realm denom always
	// starts with "/" and a native denom with a letter, so nothing needs it
	// there, and admitting it would let "5-foo" and "100-foo" quietly parse as
	// a coin whose denom is "-foo". (Amounts are unsigned, so this is a denom
	// nobody meant to write, not a parsing ambiguity.)
	//
	// Sub-realm "#" paths are absent here, which is safe only because the banker
	// refuses BankerTypeRealmIssue for sub-realms (banker.gno).
	reDnmString = `[a-z\/][a-z0-9_.:\/\-]{2,}`
	reAmt       = `[[:digit:]]+`
	reSpc       = `[[:space:]]*`
	reCoin      = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnmString))
)

// maxBaseDenomLength mirrors isValidBaseDenom in
// gnovm/stdlibs/chain/banker/banker.gno. Go cannot import a constant out of
// .gno source, so the two are kept in step by test: TestValidateDenomLength pins
// this copy, and the paired 16-byte and 17-byte base denom cases in
// gno.land/pkg/integration/testdata/realm_banker_issued_coin_denom.txtar pin
// the Gno side from both directions.
const maxBaseDenomLength = 16

// MaxDenomLength is the longest a denom may be, in bytes — 274 today, being a
// 256-byte package path, a 16-byte base name, and the two separators. The
// charset is ASCII-only, so bytes and characters are the same thing here.
//
// It is the longest denom the chain can actually produce. chain.CoinDenom in
// the Gno standard library builds a realm-issued denom as
//
//	"/" + pkgPath + ":" + baseName    e.g. "/gno.land/r/demo/foo:example"
//
// where pkgPath is capped at pkgPathLimit when the package is deployed
// (MemPackage.ValidateBasic, memfile.go) and baseName is capped at
// maxBaseDenomLength by the banker stdlib. Taking the exact sum is deliberate:
// any smaller value would let a realm deploy at a perfectly legal package path
// and then silently fail to issue coins.
//
// Those two caps only cover realm issuance, whereas ValidateDenom is the gate
// every denom passes through whatever its origin, so restating them here is
// what extends the limit to denoms arriving from a decoded transaction, a
// genesis file, or bank params.
//
// Denom bytes are not free, and for a balance held in its own store key this cap
// is the only thing bounding them, since store keys are not gas-metered. See
// tm2/pkg/sdk/bank/balance.go.
const MaxDenomLength = len("/") + pkgPathLimit + len(":") + maxBaseDenomLength

const MaxCoinsCount = 256

func ValidateDenom(denom string) error {
	// Length first: cheaper than the pattern, and it names the real problem
	// instead of a generic "invalid denom".
	if len(denom) > MaxDenomLength {
		return fmt.Errorf("denom length %d exceeds limit %d", len(denom), MaxDenomLength)
	}
	if !validDenom(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
}

// validDenom is `^reDnmString$` as a byte scan, and must stay exactly equivalent to
// the compiled form (built in coin_test.go, the only place it is still needed) —
// TestValidDenomMatchesRegexp is the gate.
//
// Hand-rolled because this is on paths where the cost is visible: banker.GetCoin and
// banker.TotalCoin validate a realm-supplied denom on every call, and the invariants
// validate every denom in state. The regexp measured 4,446ns on a maximal 274-byte
// denom against a native gas charge of a few hundred; this is 174ns.
func validDenom(denom string) bool {
	// reDnmString is `[a-z/][a-z0-9_.:/\-]{2,}`, anchored, so: at least three bytes,
	// the first from the leading class, the rest from the continuation class.
	if len(denom) < 3 {
		return false
	}
	if c := denom[0]; (c < 'a' || c > 'z') && c != '/' {
		return false
	}
	for i := 1; i < len(denom); i++ {
		switch c := denom[i]; {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '_', c == '.', c == ':', c == '/', c == '-':
		default:
			return false
		}
	}
	return true
}

// IsRealmDenom reports whether denom is one a realm may issue.
//
// This answers who may *create* a denom, not where its balance is stored — that
// is an allowlist in the bank (ViewKeeper.inAccountTier). Do not conflate them:
// an IBC voucher is neither realm-issuable nor account-tier.
//
// chain.CoinDenom builds a realm denom as "/" + pkgPath + ":" + base and
// reDnmString's leading class is [a-z\/], so a leading "/" identifies exactly
// the realm-issuable set. Enforced at SDKBanker.IssueCoin, which makes it a
// security boundary: without it a realm could mint the chain's gas denom.
func IsRealmDenom(denom string) bool {
	return strings.HasPrefix(denom, "/")
}

// ParseRealmDenom splits a realm-issued denom into its package path and base name
// and reports whether it has the shape the banker can actually issue:
// "/" + pkgPath + ":" + base.
//
// Callers must have established IsRealmDenom. This mirrors assertCoinDenom and
// isValidBaseDenom in gnovm/stdlibs/chain/banker/banker.gno — deliberately without
// their minimum base length, which is banker ergonomics with no consequence for
// stored state, and which existing realm denoms in tests do not satisfy.
//
// A realm-shaped denom that fails this is not necessarily corruption: ValidateDenom
// accepts shapes no realm could ever mint (a genesis file may name one), so a
// caller checking stored state should report rather than reject.
func ParseRealmDenom(denom string) (pkgPath, base string, err error) {
	if !IsRealmDenom(denom) {
		return "", "", fmt.Errorf("denom %q is not realm-qualified", denom)
	}
	// The package path admits no colon, so the first one separates the base name.
	pkgPath, base, ok := strings.Cut(denom[1:], ":")
	if !ok {
		return "", "", fmt.Errorf("denom %q has no base name", denom)
	}
	if len(pkgPath) > pkgPathLimit {
		return "", "", fmt.Errorf("denom %q package path is %d bytes, over the %d limit",
			denom, len(pkgPath), pkgPathLimit)
	}
	if base == "" {
		return "", "", fmt.Errorf("denom %q has an empty base name", denom)
	}
	if len(base) > maxBaseDenomLength {
		return "", "", fmt.Errorf("denom %q base name is %d bytes, over the %d limit",
			denom, len(base), maxBaseDenomLength)
	}
	if base[0] < 'a' || base[0] > 'z' {
		return "", "", fmt.Errorf("denom %q base name must start with a-z", denom)
	}
	for i := 1; i < len(base); i++ {
		c := base[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return "", "", fmt.Errorf("denom %q base name must be [a-z][a-z0-9]*", denom)
		}
	}
	return pkgPath, base, nil
}

func mustValidateDenom(denom string) {
	if err := ValidateDenom(denom); err != nil {
		panic(err)
	}
}

func MustParseCoin(coinStr string) Coin {
	coin, err := ParseCoin(coinStr)
	if err != nil {
		panic(err)
	}
	return coin
}

// ParseCoin parses a cli input for one coin type, returning errors if invalid.
// This returns an error on an empty string as well.
func ParseCoin(coinStr string) (coin Coin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	// Bound the input before the pattern runs. Coins amino-encode as a string, so
	// every transaction decode reaches here — and decode happens before the ante
	// handler installs a gas meter, so unbounded work here is unbilled. The cap is
	// MaxDenomLength plus room for the amount, which cannot exceed 19 digits.
	if len(coinStr) > MaxDenomLength+20 {
		return Coin{}, fmt.Errorf("invalid coin expression: %d bytes exceeds the limit",
			len(coinStr))
	}
	matches := reCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return Coin{}, fmt.Errorf("invalid coin expression: %s", coinStr)
	}

	denomStr, amountStr := matches[2], matches[1]

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return Coin{}, errors.Wrapf(err, "failed to parse coin amount: %s", amountStr)
	}

	if err := ValidateDenom(denomStr); err != nil {
		return Coin{}, fmt.Errorf("invalid denom cannot contain upper case characters or spaces: %w", err)
	}

	return NewCoin(denomStr, amount), nil
}

func MustParseCoins(coinsStr string) Coins {
	coins, err := ParseCoins(coinsStr)
	if err != nil {
		panic(err)
	}
	return coins
}

// ParseCoins will parse out a list of coins separated by commas.
// If nothing is provided, it returns nil Coins.
// Returned coins are sorted.
func ParseCoins(coinsStr string) (Coins, error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	coinCount := strings.Count(coinsStr, ",") + 1
	if coinCount > MaxCoinsCount {
		return nil, fmt.Errorf("coin count exceeds the limit %d", MaxCoinsCount)
	}

	coins := make(Coins, 0, coinCount)
	for coinStr := range strings.SplitSeq(coinsStr, ",") {
		coin, err := ParseCoin(coinStr)
		if err != nil {
			return nil, err
		}

		coins = append(coins, coin)
	}

	// sort coins for determinism
	coins.Sort()

	// validate coins before returning
	if err := coins.validate(); err != nil {
		return nil, fmt.Errorf("parseCoins: invalid coins %v: %w", coins, err)
	}

	return coins, nil
}

// findDup works on the assumption that coins is sorted
func findDup(coins Coins) int {
	if len(coins) <= 1 {
		return -1
	}

	prevDenom := coins[0].Denom
	for i := 1; i < len(coins); i++ {
		if coins[i].Denom == prevDenom {
			return i
		}
		prevDenom = coins[i].Denom
	}

	return -1
}
