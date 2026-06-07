// eval realm is capable of evaluating 32-bit integer
// expressions as they would appear in Go. For example:
// /r/demo/math_eval:(4+12)/2-1+11*15
package eval

import (
	evalint32 "gno.land/p/demo/math_eval/int32"
	"gno.land/p/nt/ufmt/v0"
)

func Render(p string) string {
	if len(p) == 0 {
		return `
evaluates 32-bit integer expressions. for example:
		
[(4+12)/2-1+11*15](/r/demo/math_eval:(4+12)/2-1+11*15)

`
	}
	expr, err := evalint32.Parse(p)
	if err != nil {
		return err.Error()
	}
	res, err := evalint32.Eval(expr, nil)
	if err != nil {
		return err.Error()
	}

	return ufmt.Sprintf("%s = %d", p, res)
}
