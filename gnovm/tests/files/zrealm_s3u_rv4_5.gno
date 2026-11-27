// PKGPATH: gno.land/r/zrealm_s3u_rv4_5
// R2a-save rv4: non-crossing helper writes s[0] of a caller's fresh slice and m[k] of a caller's fresh map; caller stores both
package zrealm_s3u_rv4_5

import "gno.land/r/tests/vm/s3urv"

var ss [][]string
var ms []map[string]int

func main(cur realm) {
	s := make([]string, 1)
	m := map[string]int{}
	s3urv.FillElem(s)
	s3urv.FillMap(m)
	ss = append(ss, s)
	ms = append(ms, m)
	println("owner:", ss[0][0], ms[0]["k"])
}

// Error:
// gno.land/r/tests/vm/s3urv's code was entered without crossing and is read-only (call its crossing function): s<VPBlock(1,0)>[(const (0 int))]
