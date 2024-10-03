package vm_test

import (
	"math"
	"testing"

	"github.com/gnolang/gno/gno/pkg/vm"
)

func TestStaticBlock_Define2_MaxNames(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			panicString, ok := r.(string)
			if !ok {
				t.Errorf("expected panic string, got %v", r)
			}

			if panicString != "too many variables in block" {
				t.Errorf("expected panic string to be 'too many variables in block', got '%s'", panicString)
			}

			return
		}

		// If it didn't panic, fail.
		t.Errorf("expected panic when exceeding maximum number of names")
	}()

	staticBlock := new(vm.StaticBlock)
	staticBlock.NumNames = math.MaxUint16 - 1
	staticBlock.Names = make([]vm.Name, staticBlock.NumNames)

	// Adding one more is okay.
	staticBlock.Define2(false, vm.Name("a"), vm.BoolType, vm.TypedValue{T: vm.BoolType})
	if staticBlock.NumNames != math.MaxUint16 {
		t.Errorf("expected NumNames to be %d, got %d", math.MaxUint16, staticBlock.NumNames)
	}
	if len(staticBlock.Names) != math.MaxUint16 {
		t.Errorf("expected len(Names) to be %d, got %d", math.MaxUint16, len(staticBlock.Names))
	}

	// This one should panic because the maximum number of names has been reached.
	staticBlock.Define2(false, vm.Name("a"), vm.BoolType, vm.TypedValue{T: vm.BoolType})
}
