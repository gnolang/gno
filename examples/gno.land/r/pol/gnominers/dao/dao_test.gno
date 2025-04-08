package dao

import (
	"testing"

	gm "gno.land/r/pol/gnominers"
)

func TestMerge1After(t *testing.T) {
	from := gm.GemRock
	to := gm.GemQuartz
	switch {
	case merge(float64(to - from), 10) != .5: t.Fail()
	case merge(float64(to - from), 30) != .75: t.Fail()
	case merge(float64(to - from), 100) < .9: t.Fail()
	case merge(float64(to - from), 1000) < .99: t.Fail()
	}
}

func TestMergeAlotAfter(t *testing.T) {
	from := gm.GemRock
	switch {
	case merge(float64(gm.GemQuartz - from), 10) != .5: t.Fail()
	case merge(float64(gm.GemAmethyst - from), 100) != .5: t.Fail()
	case merge(float64(gm.GemTopaz - from), 1000) != .5: t.Fail()
	case merge(float64(gm.GemAlexandrite - from), 1_000_000_000) != .5: t.Fail()
	}
}

func TestUnMerge(t *testing.T) {
	from := gm.GemQuartz
	to := gm.GemRock
	switch {
	case unMerge(float64(from - to), 1) != 10: t.Fail()
	case unMerge(float64(from - to), 2) != 20: t.Fail()
	case unMerge(float64(from - to), 10) != 100: t.Fail()
	}
}