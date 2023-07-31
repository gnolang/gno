package float64

import (
	"std"
	"testing"

	"gno.land/p/demo/ufmt"
)

func TestOne(t *testing.T) {
	ttt := []struct {
		exp string
		res float64
	}{
		{"1", 1},
		{"--1", 1},
		{"1+2", 3},
		{"-1+2", 1},
		{"-(1+2)", -3},
		{"-(1+2)*5", -15},
		{"-(1+2)*5/3", -5},
		{"1+(-(1+2)*5/3)", -4},
		{"1/3.123*4", 1 / 3.123 * 4},
	}
	for _, tc := range ttt {
		t.Run(tc.exp, func(t *testing.T) {
			exp, err := Parse(tc.exp)
			if err != nil {
				t.Errorf("%s:\n%s", tc.exp, err.Error())
			} else {
				res, errEval := Eval(exp, nil)
				if errEval != nil {
					t.Errorf("eval error: %s", errEval.Error())
				} else if res != tc.res {
					t.Errorf("%s:\nexpected %d, got %d", tc.exp, tc.res, res)
				}
			}
		})
	}
}

func TestVariables(t *testing.T) {
	fn := func(x, y float64) float64 {
		return 1 + x + y
	}
	expr := "1 + x + y"
	exp, err := Parse(expr)
	if err != nil {
		t.Errorf("could not parse: %s", err.Error())
	}
	variables := make(map[string]float64)
	for i := 0.0; i < 10; i++ {
		variables["x"] = float64(i)
		variables["y"] = 2.0
		res, errEval := Eval(exp, variables)
		if errEval != nil {
			t.Errorf("could not evaluate: %s", err.Error())
		}
		expected := fn(variables["x"], variables["y"])
		if res != expected {
			t.Errorf("expected: %d, actual: %d", expected, res)
		}
	}
}
