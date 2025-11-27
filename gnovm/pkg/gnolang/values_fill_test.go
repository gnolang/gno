package gnolang

import (
	"testing"
)

func TestConstantValuesDeepFill(t *testing.T) {
	sv := StringValue("test")
	if result := sv.DeepFill(nil); result != sv {
		t.Errorf("StringValue.DeepFill: expected %v, got %v", sv, result)
	}

	biv := BigintValue{}
	if result := biv.DeepFill(nil); result != biv {
		t.Errorf("BigintValue.DeepFill: expected %v, got %v", biv, result)
	}

	bdv := BigdecValue{}
	if result := bdv.DeepFill(nil); result != bdv {
		t.Errorf("BigdecValue.DeepFill: expected %v, got %v", bdv, result)
	}
}
