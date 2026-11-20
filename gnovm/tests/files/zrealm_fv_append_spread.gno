// PKGPATH: gno.land/r/fvappendspread
package fvappendspread

import "gno.land/r/tests/vm/fvx"

var saved []fvx.T

func main(cur realm) {
	s := fvx.Sl
	saved = append(s[:0:0], s...)
	println(len(saved), saved[0].N)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:6:5: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:9:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:9:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:10:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:10:10: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:10:17: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:10:17: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvappendspread/zrealm_fv_append_spread.gno:10:26: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
