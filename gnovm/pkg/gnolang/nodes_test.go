package gnolang_test

import (
	"math"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
)

func TestStaticBlock_Define2_MaxNames(t *testing.T) {
	staticBlock := new(gnolang.StaticBlock)
	staticBlock.NumNames = math.MaxUint16 - 1
	staticBlock.Names = make([]gnolang.Name, staticBlock.NumNames)
	staticBlock.Types = make([]gnolang.Type, staticBlock.NumNames)
	staticBlock.NameSources = make([]gnolang.NameSource, staticBlock.NumNames)

	// Adding one more is okay.
	staticBlock.Define2(false, gnolang.Name("a"), gnolang.BoolType, gnolang.TypedValue{T: gnolang.BoolType}, gnolang.NameSource{})
	if staticBlock.NumNames != math.MaxUint16 {
		t.Errorf("expected NumNames to be %d, got %d", math.MaxUint16, staticBlock.NumNames)
	}
	if len(staticBlock.Names) != math.MaxUint16 {
		t.Errorf("expected len(Names) to be %d, got %d", math.MaxUint16, len(staticBlock.Names))
	}

	// This one should panic because the maximum number of names has been reached.
	assert.PanicsWithValue(
		t,
		"too many variables in block",
		func() {
			staticBlock.Define2(false, gnolang.Name("a"), gnolang.BoolType, gnolang.TypedValue{T: gnolang.BoolType}, gnolang.NameSource{})
		},
	)
}
