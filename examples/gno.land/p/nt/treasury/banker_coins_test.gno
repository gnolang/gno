package treasury

import (
	"std"
	"testing"
)

var testRealm = std.NewCodeRealm("gno.land/r/test/test")

func TestNewCoinsBanker(t *testing.T) {
	// testing.SetRealm(testRealm)
	// banker := std.NewBanker(std.BankerTypeReadonly)
	// current := testRealm.Address()
	//
	// // Test working cases.
	// b, err := NewCoinsBanker(banker)
	// urequire.NoError(t, err)
	// uassert.Equal(t, b.Address(), string(current))

	// b, err = NewCoinsBankerWithOwner(current, banker)
	// uassert.NotNil(t, b)
	// uassert.Nil(t, err)
	// uassert.Equal(t, b.Address(), string(current))
	//
	// // Test failing cases.
	// b, err = NewCoinsBankerWithOwner(std.Address(""), banker)
	// uassert.Nil(t, b)
	// uassert.NotNil(t, err)
	//
	// b, err = NewCoinsBankerWithOwner(current, nil)
	// uassert.Nil(t, b)
	// uassert.NotNil(t, err)
}
