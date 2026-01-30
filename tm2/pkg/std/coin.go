package std

import (
	"fmt"
	"regexp"
	"slices"
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
	if err := validate(denom, amount); err != nil {
		panic(err)
	}

	return Coin{
		Denom:  denom,
		Amount: amount,
	}
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
	// remove zeroes
	newCoins := removeZeroCoins(Coins(coins))
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

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin.String())
	}
	return out[:len(out)-1]
}

// IsValid asserts the Coins are sorted, have positive amount,
// and Denom does not contain upper case characters.
func (coins Coins) IsValid() bool {
	return coins.validate() == nil
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
// TODO: Remove once unsigned integers are used.
func (coins Coins) negative() Coins {
	res := make([]Coin, 0, len(coins))

	for _, coin := range coins {
		res = append(res, Coin{
			Denom:  coin.Denom,
			Amount: -1 * coin.Amount,
		})
	}

	return res
}

// removeZeroCoins removes all zero coins from the given coin set in-place.
func removeZeroCoins(coins Coins) Coins {
	i, l := 0, len(coins)
	for i < l {
		if coins[i].IsZero() {
			// remove coin
			coins = slices.Delete(coins, i, i+1)
			l--
		} else {
			i++
		}
	}

	return coins[:i]
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
	reDnmString = `[a-z\/][a-z0-9_.:\/]{2,}`
	reAmt       = `[[:digit:]]+`
	reSpc       = `[[:space:]]*`
	reDnm       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reDnmString))
	reCoin      = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnmString))
)

func ValidateDenom(denom string) error {
	if !reDnm.MatchString(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
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

	coinStrs := strings.Split(coinsStr, ",")
	coins := make(Coins, len(coinStrs))
	for i, coinStr := range coinStrs {
		coin, err := ParseCoin(coinStr)
		if err != nil {
			return nil, err
		}

		coins[i] = coin
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
