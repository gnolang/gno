package gnolang

import (
	"math/big"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// This test exposes a panics that occurs when BigintValue is embedded
// in an other struct.
func TestAminoMustMarshalJSONPanics(t *testing.T) {
	bi := BigintValue{V: big.NewInt(20)}
	b := amino.MustMarshalJSON(bi) // works well
	println(string(b))

	pv := PackageValue{
		Block: bi,
	}
	b = amino.MustMarshalJSON(pv) // panics
	println(string(b))
}
