// PKGPATH: gno.land/r/valueban2
package valueban2

import "gno.land/r/tests/vm/valueban"

// Aliases, multi-value results and comma-ok forms are judged; reading a
// length, an element's field, a range index or a discarded result is not.
var held valueban.CouponA

func main(cur realm) {
	held = valueban.One()
	_, n := valueban.Pair()
	v, ok := valueban.Boxed().(valueban.Coupon)
	println(held.Value, n, v.Value, ok)
	println(len(valueban.Arr), valueban.Arr[1].Value)
	for i := range valueban.Arr {
		println(i)
	}
	defer valueban.One()
}

// Error:
// cannot allocate gno.land/r/tests/vm/valueban.Coupon in realm gno.land/r/valueban2

// TypeCheckError:
// gno.land/r/valueban2/zrealm_value_ban2.gno:8:5: cannot hold a value of gno.land/r/tests/vm/valueban.Coupon outside its realm (use a pointer); gno.land/r/valueban2/zrealm_value_ban2.gno:11:2: cannot hold a value of gno.land/r/tests/vm/valueban.Coupon outside its realm (use a pointer); gno.land/r/valueban2/zrealm_value_ban2.gno:11:9: cannot hold a value of gno.land/r/tests/vm/valueban.Coupon outside its realm (use a pointer); gno.land/r/valueban2/zrealm_value_ban2.gno:12:10: cannot hold a value of gno.land/r/tests/vm/valueban.Coupon outside its realm (use a pointer); gno.land/r/valueban2/zrealm_value_ban2.gno:13:2: cannot hold a value of gno.land/r/tests/vm/valueban.Coupon outside its realm (use a pointer); gno.land/r/valueban2/zrealm_value_ban2.gno:13:11: cannot hold a value of gno.land/r/tests/vm/valueban.Coupon outside its realm (use a pointer)
