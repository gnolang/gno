package cover

import (
	"testing"
)

func TestIsEven(t *testing.T){
	if !IsEven(2){
		t.Error("should be even")
	}

	if !IsEven(4){
		t.Error("should be even")
	}

	if !IsEven(6){
		t.Error("should be even")
	}

	if IsEven(3){
		t.Error("should not be even")
	}

	if IsEven(3){
		t.Error("should not be even")
	}
}