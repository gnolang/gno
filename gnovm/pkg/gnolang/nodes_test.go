package gnolang_test

import (
	"fmt"
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

func TestStaticBlock_Reserve_MaxNames(t *testing.T) {
	// Reserve appends through defineNew, not Define2, so it needs its own
	// max-names guard; without it NumNames wraps around silently.
	staticBlock := new(gnolang.StaticBlock)
	staticBlock.NumNames = math.MaxUint16
	staticBlock.Names = make([]gnolang.Name, staticBlock.NumNames)
	staticBlock.Types = make([]gnolang.Type, staticBlock.NumNames)
	staticBlock.NameSources = make([]gnolang.NameSource, staticBlock.NumNames)
	staticBlock.HeapItems = make([]bool, staticBlock.NumNames)

	assert.PanicsWithValue(
		t,
		"too many variables in block",
		func() {
			staticBlock.Reserve(false, &gnolang.NameExpr{Name: "a"}, nil, gnolang.NSDefine, 0)
		},
	)
}

func TestStaticBlock_GetLocalIndex_LastWinsPastThreshold(t *testing.T) {
	// A faux case block that shadows a name copied in from its if/switch
	// holds that name twice, and GetLocalIndex must answer the later slot.
	// Past nameIndexThreshold (32) that answer comes from nameIndex rather
	// than the linear scan, so buildNameIndex has to be last-wins too.
	// Both widths are checked so the two branches cannot drift apart.
	for _, width := range []int{20, 40} {
		sb := new(gnolang.StaticBlock)
		for i := range width {
			n := gnolang.Name(fmt.Sprintf("n%02d", i))
			switch i {
			case 0:
				n = "v" // the faux copy
			case 3:
				n = "v" // the shadow declared in the case body
			}
			sb.Names = append(sb.Names, n)
			sb.Types = append(sb.Types, nil)
			sb.NameSources = append(sb.NameSources, gnolang.NameSource{})
			sb.HeapItems = append(sb.HeapItems, false)
		}
		sb.NumNames = uint16(width)

		idx, ok := sb.GetLocalIndex("v")
		assert.True(t, ok, "width %d", width)
		assert.Equal(t, uint16(3), idx, "width %d: want the shadow's slot", width)
	}
}
