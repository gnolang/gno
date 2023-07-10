// Microblog is a website with shortform posts from users.
// The API is simple - "AddPost" takes markdown and
// adds it to the users site.
// The microblog location is determined by the user address
// /r/demo/microblog:<YOUR-ADDRESS>
package eval

import (
	evalint32 "gno.land/p/demo/eval/int32"
	"gno.land/p/demo/ufmt"
)

func Render(p string) string {
	if len(p) == 0 {
		return `
evaluates 32-bit integer expressions. for example:
		
[(4+12)/2-1+11*15](/r/demo/eval:(4+12)/2-1+11*15)

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
