package shop

import "chain/runtime/unsafe"

var collected int64

func Buy(cur realm) {
	if !cur.Previous().IsUser() {
		panic("user only")
	}
	collected += unsafe.OriginSend().AmountOf("ugnot")
}

func Collected() int64 {
	return collected
}
