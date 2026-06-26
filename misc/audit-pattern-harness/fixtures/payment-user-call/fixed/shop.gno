package shop

import "chain/runtime/unsafe"

var collected int64

func Buy(cur realm) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	if !cur.Previous().IsUserCall() {
		panic("direct user call only")
	}
	collected += unsafe.OriginSend().AmountOf("ugnot")
}

func Collected() int64 {
	return collected
}
