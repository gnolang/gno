package int32

import (
	"std"
	"testing"

	"gno.land/p/demo/ufmt"
)

func TestOne(t *testing.T) {
	ttt := []struct {
		exp string
		res int
	}{
		{"1", 1},
		{"--1", 1},
		{"1+2", 3},
		{"-1+2", 1},
		{"-(1+2)", -3},
		{"-(1+2)*5", -15},
		{"-(1+2)*5/3", -5},
		{"1+(-(1+2)*5/3)", -4},
		{"3^4", 3 ^ 4},
		{"8%2", 8 % 2},
		{"8%3", 8 % 3},
		{"8|3", 8 | 3},
		{"10%2", 0},
		{"(4    + 3)/2-1+11*15", (4+3)/2 - 1 + 11*15},
		{
			"(30099>>10^30099>>11)%5*((30099>>14&3^30099>>15&1)+1)*30099%99 + ((3 + (30099 >> 14 & 3) - (30099 >> 16 & 1)) / 3 * 30099 % 99 & 64)",
			(30099>>10^30099>>11)%5*((30099>>14&3^30099>>15&1)+1)*30099%99 + ((3 + (30099 >> 14 & 3) - (30099 >> 16 & 1)) / 3 * 30099 % 99 & 64),
		},
		{
			"(1023850>>10^1023850>>11)%5*((1023850>>14&3^1023850>>15&1)+1)*1023850%99 + ((3 + (1023850 >> 14 & 3) - (1023850 >> 16 & 1)) / 3 * 1023850 % 99 & 64)",
			(1023850>>10^1023850>>11)%5*((1023850>>14&3^1023850>>15&1)+1)*1023850%99 + ((3 + (1023850 >> 14 & 3) - (1023850 >> 16 & 1)) / 3 * 1023850 % 99 & 64),
		},
		{"((0000+1)*0000)", 0},
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
	fn := func(x, y int) int {
		return 1 + ((x*3+1)*(x*2))>>y + 1
	}
	expr := "1 + ((x*3+1)*(x*2))>>y + 1"
	exp, err := Parse(expr)
	if err != nil {
		t.Errorf("could not parse: %s", err.Error())
	}
	variables := make(map[string]int)
	for i := 0; i < 10; i++ {
		variables["x"] = i
		variables["y"] = 2
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
