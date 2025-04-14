package gnominers

import (
	"testing"
)

func TestMerge1After(t *testing.T) {
	from := gemRock
	to := gemQuartz
	switch {
	case merge(float64(to-from), 10) != .5:
		t.Fail()
	case merge(float64(to-from), 30) != .75:
		t.Fail()
	case merge(float64(to-from), 100) < .9:
		t.Fail()
	case merge(float64(to-from), 1000) < .99:
		t.Fail()
	}
}

func TestMergeAlotAfter(t *testing.T) {
	from := gemRock
	switch {
	case merge(float64(gemQuartz-from), 10) != .5:
		t.Fail()
	case merge(float64(gemAmethyst-from), 100) != .5:
		t.Fail()
	case merge(float64(gemTopaz-from), 1000) != .5:
		t.Fail()
	case merge(float64(gemAlexandrite-from), 1_000_000_000) != .5:
		t.Fail()
	}
}

func TestUnMerge(t *testing.T) {
	from := gemQuartz
	to := gemRock
	switch {
	case unMerge(float64(from-to), 1) != 10:
		t.Fail()
	case unMerge(float64(from-to), 2) != 20:
		t.Fail()
	case unMerge(float64(from-to), 10) != 100:
		t.Fail()
	}
}
