// Package valueban backs the realm value rule filetests.
package valueban

type Coupon struct{ Value int }

type CouponA = Coupon

var Arr = [2]Coupon{{1}, {2}}

func One() CouponA        { return Coupon{3} }
func Pair() (Coupon, int) { return Coupon{5}, 5 }
func Boxed() any          { return Coupon{7} }
