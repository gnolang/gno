package std

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDenom1       = "atom"
	testDenom2       = "muon"
	testDenomInvalid = "Atom"
)

// ----------------------------------------------------------------------------
// Coin tests

func TestCoin(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() { NewCoin(testDenom1, -1) })
	require.Panics(t, func() { NewCoin(strings.ToUpper(testDenom1), 10) })
	require.Equal(t, int64(5), NewCoin(testDenom1, 5).Amount)
}

func TestIsEqualCoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 1), true, false},
		{NewCoin(testDenom1, 1), NewCoin(testDenom2, 1), false, true},
		{NewCoin("stake", 1), NewCoin("stake", 10), false, false},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin equality relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoinIsValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		coin       Coin
		expectPass bool
	}{
		{Coin{testDenom1, int64(-1)}, false},
		{Coin{testDenom1, int64(0)}, true},
		{Coin{testDenom1, int64(1)}, true},
		{Coin{"Atom", int64(1)}, false},
		{Coin{"a", int64(1)}, false},
		{Coin{"a very long coin denom", int64(1)}, false},
		{Coin{"atOm", int64(1)}, false},
		{Coin{"     ", int64(1)}, false},
	}

	for i, tc := range cases {
		require.Equal(t, tc.expectPass, tc.coin.IsValid(), "unexpected result for IsValid, tc #%d", i)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		denom   string
		amount  int64
		wantErr string
	}{
		{"atom", 1, ""},
		{"atom", 0, ""},
		{"atom", -1, "negative coin amount: -1"},
		{"Atom", 1, "invalid denom: Atom"},
		{"a", 1, "invalid denom: a"},
		{"a very long coin denom", 1, "invalid denom: a very long coin denom"},
		{"atOm", 1, "invalid denom: atOm"},
		{"     ", 1, "invalid denom:      "},
	}

	for i, tc := range cases {
		err := validate(tc.denom, tc.amount)
		if tc.wantErr == "" {
			require.NoError(t, err, "unexpected error for validate, tc #%d", i)
		} else {
			require.EqualError(t, err, tc.wantErr, "unexpected error message for validate, tc #%d", i)
		}
	}
}

func TestCoinsValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		coins   Coins
		wantErr string
	}{
		{Coins{}, ""},
		{Coins{{"gas", 1}}, ""},
		{Coins{{"gas", 1}, {"mineral", 1}}, ""},
		{Coins{{"gas", 0}}, "non-positive coin amount: 0"},
		{Coins{{"gas", -1}}, "non-positive coin amount: -1"},
		{Coins{{"GAS", 1}}, "invalid denom: GAS"},
		{Coins{{"gas", 1}, {"MINERAL", 1}}, "invalid denom: MINERAL"},
		{Coins{{"bbb", 1}, {"aaa", 1}}, "coins not sorted: aaa < bbb"},
		{Coins{{"gas", 1}, {"tree", 1}, {"mineral", 1}}, "coins not sorted: mineral < tree"},
		{Coins{{"gas", 1}, {"gas", 1}}, "duplicate denom: gas"},
		{Coins{{"gas", 1}, {"mineral", 1}, {"mineral", 1}}, "duplicate denom: mineral"},
		{Coins{{"gas", 1}, {"mineral", 0}}, "non-positive coin amount: 0"},
		{Coins{{"gas", 1}, {"mineral", -5}}, "non-positive coin amount: -5"},
	}

	for i, tc := range cases {
		err := tc.coins.validate()
		if tc.wantErr == "" {
			require.NoError(t, err, "unexpected error for Coins.validate, tc #%d", i)
		} else {
			require.EqualError(t, err, tc.wantErr, "unexpected error message for Coins.validate, tc #%d", i)
		}
	}
}

func TestAddCoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 1), NewCoin(testDenom1, 2), false},
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 0), NewCoin(testDenom1, 1), false},
		{NewCoin(testDenom1, 1), NewCoin(testDenom2, 1), NewCoin(testDenom1, 1), true},
		{Coin{Denom: testDenomInvalid, Amount: 1}, Coin{Denom: testDenomInvalid, Amount: 1}, NewCoin(testDenom1, 0), true},
	}

	for tcIndex, tc := range cases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Add(tc.inputTwo) })
		} else {
			res := tc.inputOne.Add(tc.inputTwo)
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSubCoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewCoin(testDenom1, 1), NewCoin(testDenom2, 1), NewCoin(testDenom1, 1), true},
		{NewCoin(testDenom1, 10), NewCoin(testDenom1, 1), NewCoin(testDenom1, 9), false},
		{NewCoin(testDenom1, 5), NewCoin(testDenom1, 3), NewCoin(testDenom1, 2), false},
		{NewCoin(testDenom1, 5), NewCoin(testDenom1, 0), NewCoin(testDenom1, 5), false},
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 5), Coin{}, true},
		{NewCoin(testDenom1, 1), Coin{Denom: testDenomInvalid, Amount: 1}, Coin{}, true},
	}

	for tcIndex, tc := range cases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			require.Equal(t, tc.expected, res, "difference of coins is incorrect, tc #%d", tcIndex)
		}
	}

	tc := struct {
		inputOne Coin
		inputTwo Coin
		expected int64
	}{NewCoin(testDenom1, 1), NewCoin(testDenom1, 1), 0}
	res := tc.inputOne.Sub(tc.inputTwo)
	require.Equal(t, tc.expected, res.Amount)
}

func TestIsGTECoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 1), true, false},
		{NewCoin(testDenom1, 2), NewCoin(testDenom1, 1), true, false},
		{NewCoin(testDenom1, 1), NewCoin(testDenom2, 1), false, true},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsGTE(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsGTE(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin GTE relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestIsLTCoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 1), false, false},
		{NewCoin(testDenom1, 2), NewCoin(testDenom1, 1), false, false},
		{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1), false, true},
		{NewCoin(testDenom1, 1), NewCoin(testDenom2, 1), false, true},
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 1), false, false},
		{NewCoin(testDenom1, 1), NewCoin(testDenom1, 2), true, false},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsLT(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsLT(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin LT relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoinIsZero(t *testing.T) {
	t.Parallel()

	coin := NewCoin(testDenom1, 0)
	res := coin.IsZero()
	require.True(t, res)

	coin = NewCoin(testDenom1, 1)
	res = coin.IsZero()
	require.False(t, res)
}

// ----------------------------------------------------------------------------
// Coins tests

func TestIsZeroCoins(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne Coins
		expected bool
	}{
		{Coins{}, true},
		{Coins{NewCoin(testDenom1, 0)}, true},
		{Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 0)}, true},
		{Coins{NewCoin(testDenom1, 1)}, false},
		{Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1)}, false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsZero()
		require.Equal(t, tc.expected, res)
	}
}

func TestEqualCoins(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected bool
		panics   bool
	}{
		{Coins{}, Coins{}, true, false},
		{Coins{NewCoin(testDenom1, 0)}, Coins{NewCoin(testDenom1, 0)}, true, false},
		{Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1)}, Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1)}, true, false},
		{Coins{NewCoin(testDenom1, 0)}, Coins{NewCoin(testDenom2, 0)}, false, true},
		{Coins{NewCoin(testDenom1, 0)}, Coins{NewCoin(testDenom1, 1)}, false, false},
		{Coins{NewCoin(testDenom1, 0)}, Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1)}, false, false},
		{Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1)}, Coins{NewCoin(testDenom1, 0), NewCoin(testDenom2, 1)}, true, false},
	}

	for tcnum, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			require.Equal(t, tc.expected, res, "Equality is differed from exported. tc #%d, expected %b, actual %b.", tcnum, tc.expected, res)
		}
	}
}

func TestAddCoins(t *testing.T) {
	t.Parallel()

	zero := int64(0)
	one := int64(1)
	two := int64(2)

	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected Coins
		panics   bool
	}{
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}, {testDenom2, two}}, false},
		{Coins{{testDenom1, zero}, {testDenom2, one}}, Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins{{testDenom2, one}}, false},
		{Coins{{testDenom1, two}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, two}}, false},
		{Coins{{testDenom1, one}}, Coins{{testDenom1, one}, {testDenom2, two}}, Coins{{testDenom1, two}, {testDenom2, two}}, false},
		{Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins(nil), false},
		{Coins{{testDenom1, zero}}, Coins{{testDenomInvalid, one}}, Coins{}, true},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.Add(tc.inputTwo) })
		} else {
			res := tc.inputOne.Add(tc.inputTwo)
			assert.True(t, res.IsValid())
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSubCoins(t *testing.T) {
	t.Parallel()

	zero := int64(0)
	one := int64(1)
	two := int64(2)

	testCases := []struct {
		inputOne    Coins
		inputTwo    Coins
		expected    Coins
		shouldPanic bool
	}{
		{Coins{{testDenom1, two}}, Coins{{testDenom1, one}, {testDenom2, two}}, Coins{{testDenom1, one}, {testDenom2, two}}, true},
		{Coins{{testDenom1, two}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, two}}, false},
		{Coins{{testDenom1, one}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, one}}, false},
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}}, Coins{{testDenom2, one}}, false},
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}}, Coins{}, true},
	}

	for i, tc := range testCases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			assert.True(t, res.IsValid())
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", i)
		}
	}
}

func TestCoins(t *testing.T) {
	t.Parallel()

	good := Coins{
		{"gas", int64(1)},
		{"mineral", int64(1)},
		{"tree", int64(1)},
	}
	mixedCase1 := Coins{
		{"gAs", int64(1)},
		{"MineraL", int64(1)},
		{"TREE", int64(1)},
	}
	mixedCase2 := Coins{
		{"gAs", int64(1)},
		{"mineral", int64(1)},
	}
	mixedCase3 := Coins{
		{"gAs", int64(1)},
	}
	empty := NewCoins()
	badSort1 := Coins{
		{"tree", int64(1)},
		{"gas", int64(1)},
		{"mineral", int64(1)},
	}

	// both are after the first one, but the second and third are in the wrong order
	badSort2 := Coins{
		{"gas", int64(1)},
		{"tree", int64(1)},
		{"mineral", int64(1)},
	}
	badAmt := Coins{
		{"gas", int64(1)},
		{"tree", int64(0)},
		{"mineral", int64(1)},
	}
	dup := Coins{
		{"gas", int64(1)},
		{"gas", int64(1)},
		{"mineral", int64(1)},
	}
	neg := Coins{
		{"gas", int64(-1)},
		{"mineral", int64(1)},
	}

	assert.True(t, good.IsValid(), "Coins are valid")
	assert.False(t, mixedCase1.IsValid(), "Coins denoms contain upper case characters")
	assert.False(t, mixedCase2.IsValid(), "First Coins denoms contain upper case characters")
	assert.False(t, mixedCase3.IsValid(), "Single denom in Coins contains upper case characters")
	assert.True(t, good.IsAllPositive(), "Expected coins to be positive: %v", good)
	assert.False(t, empty.IsAllPositive(), "Expected coins to not be positive: %v", empty)
	assert.True(t, good.IsAllGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(t, good.IsAllLT(empty), "Expected %v to be < %v", good, empty)
	assert.True(t, empty.IsAllLT(good), "Expected %v to be < %v", empty, good)
	assert.False(t, badSort1.IsValid(), "Coins are not sorted")
	assert.False(t, badSort2.IsValid(), "Coins are not sorted")
	assert.False(t, badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(t, dup.IsValid(), "Duplicate coin")
	assert.False(t, neg.IsValid(), "Negative first-denom coin")
}

func TestCoinsGT(t *testing.T) {
	t.Parallel()

	one := int64(1)
	two := int64(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom2, two}}))
}

func TestCoinsLT(t *testing.T) {
	t.Parallel()

	one := int64(1)
	two := int64(2)

	assert.False(t, Coins{}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, two}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLT(Coins{{testDenom1, one}}))
}

func TestCoinsLTE(t *testing.T) {
	t.Parallel()

	one := int64(1)
	two := int64(2)

	assert.True(t, Coins{}.IsAllLTE(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLTE(Coins{{testDenom1, one}}))
}

func TestParse(t *testing.T) {
	t.Parallel()

	one := int64(1)

	cases := []struct {
		input    string
		valid    bool  // if false, we expect an error on parse
		expected Coins // if valid is true, make sure this is returned
	}{
		{"", true, nil},
		{"1foo", true, Coins{{"foo", one}}},
		{"10bar", true, Coins{{"bar", int64(10)}}},
		{"99bar,1foo", true, Coins{{"bar", int64(99)}, {"foo", one}}},
		{"98 bar , 1 foo  ", true, Coins{{"bar", int64(98)}, {"foo", one}}},
		{"  55\t \t bling\n", true, Coins{{"bling", int64(55)}}},
		{"2foo, 97 bar", true, Coins{{"bar", int64(97)}, {"foo", int64(2)}}},
		{"5foo-bar", false, nil},
		{"5 mycoin,", false, nil},             // no empty coins in a list
		{"2 3foo, 97 bar", false, nil},        // 3foo is invalid coin name
		{"11me coin, 12you coin", false, nil}, // no spaces in coin names
		{"1.2btc", false, nil},                // amount must be integer
		{"-5foo", false, nil},                 // amount must be positive
		{"5Foo", false, nil},                  // denom must be lowercase
	}

	for tcIndex, tc := range cases {
		res, err := ParseCoins(tc.input)
		if !tc.valid {
			require.NotNil(t, err, "%s: %#v. tc #%d", tc.input, res, tcIndex)
		} else if assert.Nil(t, err, "%s: %+v", tc.input, err) {
			require.Equal(t, tc.expected, res, "coin parsing was incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSortCoins(t *testing.T) {
	t.Parallel()

	good := Coins{
		NewCoin("gas", 1),
		NewCoin("mineral", 1),
		NewCoin("tree", 1),
	}
	empty := Coins{
		NewCoin("gold", 0),
	}
	badSort1 := Coins{
		NewCoin("tree", 1),
		NewCoin("gas", 1),
		NewCoin("mineral", 1),
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		NewCoin("gas", 1),
		NewCoin("tree", 1),
		NewCoin("mineral", 1),
	}
	badAmt := Coins{
		NewCoin("gas", 1),
		NewCoin("tree", 0),
		NewCoin("mineral", 1),
	}
	dup := Coins{
		NewCoin("gas", 1),
		NewCoin("gas", 1),
		NewCoin("mineral", 1),
	}

	cases := []struct {
		coins         Coins
		before, after bool // valid before/after sort
	}{
		{good, true, true},
		{empty, false, false},
		{badSort1, false, true},
		{badSort2, false, true},
		{badAmt, false, false},
		{dup, false, false},
	}

	for tcIndex, tc := range cases {
		require.Equal(t, tc.before, tc.coins.IsValid(), "coin validity is incorrect before sorting, tc #%d", tcIndex)
		tc.coins.Sort()
		require.Equal(t, tc.after, tc.coins.IsValid(), "coin validity is incorrect after sorting, tc #%d", tcIndex)
	}
}

func TestAmountOf(t *testing.T) {
	t.Parallel()

	case0 := Coins{}
	case1 := Coins{
		NewCoin("gold", 0),
	}
	case2 := Coins{
		NewCoin("gas", 1),
		NewCoin("mineral", 1),
		NewCoin("tree", 1),
	}
	case3 := Coins{
		NewCoin("mineral", 1),
		NewCoin("tree", 1),
	}
	case4 := Coins{
		NewCoin("gas", 8),
	}

	cases := []struct {
		coins           Coins
		amountOf        int64
		amountOfSpace   int64
		amountOfGAS     int64
		amountOfMINERAL int64
		amountOfTREE    int64
	}{
		{case0, 0, 0, 0, 0, 0},
		{case1, 0, 0, 0, 0, 0},
		{case2, 0, 0, 1, 1, 1},
		{case3, 0, 0, 0, 1, 1},
		{case4, 0, 0, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.amountOfGAS, tc.coins.AmountOf("gas"))
		assert.Equal(t, tc.amountOfMINERAL, tc.coins.AmountOf("mineral"))
		assert.Equal(t, tc.amountOfTREE, tc.coins.AmountOf("tree"))
	}

	assert.Panics(t, func() { cases[0].coins.AmountOf("Invalid") })
}

func TestCoinsIsAnyGTE(t *testing.T) {
	t.Parallel()

	one := int64(1)
	two := int64(2)

	assert.False(t, Coins{}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{"xxx", one}, {"yyy", one}}.IsAnyGTE(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestCoinsIsAllGT(t *testing.T) {
	t.Parallel()

	one := int64(1)
	two := int64(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, two}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{"xxx", one}, {"yyy", one}}.IsAllGT(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestCoinsIsAllGTE(t *testing.T) {
	t.Parallel()

	one := int64(1)
	two := int64(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, two}}))
	assert.False(t, Coins{}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, two}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{"xxx", one}, {"yyy", one}}.IsAllGTE(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestNewCoins(t *testing.T) {
	t.Parallel()

	tenatom := NewCoin("atom", 10)
	tenbtc := NewCoin("btc", 10)
	zeroeth := NewCoin("eth", 0)

	// don't use NewCoin(...) to avoid early panic
	uppercase := Coin{
		Denom:  "UPC",
		Amount: 10,
	}
	negative := Coin{
		Denom:  "neg",
		Amount: -5,
	}

	tests := []struct {
		name      string
		coins     Coins
		want      Coins
		wantPanic bool
	}{
		{"empty args", []Coin{}, Coins{}, false},
		{"one coin", []Coin{tenatom}, Coins{tenatom}, false},
		{"sort after create", []Coin{tenbtc, tenatom}, Coins{tenatom, tenbtc}, false},
		{"sort and remove zeroes", []Coin{zeroeth, tenbtc, tenatom}, Coins{tenatom, tenbtc}, false},
		{"panic on uppercase denom", []Coin{uppercase}, Coins{}, true},
		{"panic on negative amount", []Coin{negative}, Coins{}, true},
		{"panic on dups", []Coin{tenatom, tenatom}, Coins{}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				require.Panics(t, func() { NewCoins(tt.coins...) })
				return
			}
			got := NewCoins(tt.coins...)
			require.True(t, got.IsEqual(tt.want))
		})
	}
}

func TestCoinsIsAnyGT(t *testing.T) {
	t.Parallel()

	twoAtom := NewCoin("atom", 2)
	fiveAtom := NewCoin("atom", 5)
	threeEth := NewCoin("eth", 3)
	sixEth := NewCoin("eth", 6)
	twoBtc := NewCoin("btc", 2)

	require.False(t, Coins{}.IsAnyGT(Coins{}))

	require.False(t, Coins{fiveAtom}.IsAnyGT(Coins{}))
	require.False(t, Coins{}.IsAnyGT(Coins{fiveAtom}))
	require.True(t, Coins{fiveAtom}.IsAnyGT(Coins{twoAtom}))
	require.False(t, Coins{twoAtom}.IsAnyGT(Coins{fiveAtom}))

	require.True(t, Coins{twoAtom, sixEth}.IsAnyGT(Coins{twoBtc, fiveAtom, threeEth}))
	require.False(t, Coins{twoBtc, twoAtom, threeEth}.IsAnyGT(Coins{fiveAtom, sixEth}))
	require.False(t, Coins{twoAtom, sixEth}.IsAnyGT(Coins{twoBtc, fiveAtom}))
}

func TestFindDup(t *testing.T) {
	t.Parallel()

	abc := NewCoin("abc", 10)
	def := NewCoin("def", 10)
	ghi := NewCoin("ghi", 10)

	type args struct {
		coins Coins
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{NewCoins()}, -1},
		{"one coin", args{NewCoins(NewCoin("xyz", 10))}, -1},
		{"no dups", args{Coins{abc, def, ghi}}, -1},
		{"dup at first position", args{Coins{abc, abc, def}}, 1},
		{"dup after first position", args{Coins{abc, def, def}}, 2},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := findDup(tt.args.coins); got != tt.want {
				t.Errorf("findDup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalJSONCoins(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     Coins
		strOutput string
	}{
		{"nil coins", nil, `""`},
		{"empty coins", Coins{}, `""`},
		{"non-empty coins", NewCoins(NewCoin("foo", 50)), `"50foo"`},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bz, err := amino.MarshalJSON(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.strOutput, string(bz))

			var newCoins Coins
			require.NoError(t, amino.UnmarshalJSON(bz, &newCoins))

			if tc.input.Empty() {
				require.Nil(t, newCoins)
			} else {
				require.Equal(t, tc.input, newCoins)
			}
		})
	}
}

func TestContainOneOfDenom(t *testing.T) {
	restrictList := map[string]struct{}{
		"baz": {},
		"foo": {},
	}
	amt := Coins{
		{"foo", int64(1)},
		{"bar", int64(1)},
	}
	require.True(t, amt.ContainOneOfDenom(restrictList))

	zero := Coins{
		{"foo", int64(0)},
		{"bar", int64(1)},
	}

	// only return true when the value is posible
	require.False(t, zero.ContainOneOfDenom(restrictList))
}
