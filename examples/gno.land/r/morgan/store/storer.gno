package store // r/store

import "gno.land/p/moul/deque"

var stored *deque.Deque

func PushAndStore(cur realm, dq *deque.Deque) {
	dq.PushBack("123")
	stored = dq
}
