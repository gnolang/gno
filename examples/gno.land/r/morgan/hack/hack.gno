package hack

import (
	"strconv"

	"gno.land/p/moul/deque"
	"gno.land/r/morgan/store"
)

var mine *deque.Deque

func Call(cur realm) {
	dq := deque.New()
	dq.PushBack("000")
	(func(cur realm) { mine = dq })(cross)
	store.PushAndStore(cross, dq)
	dq.PushBack("888")
}

func Render(string) string {
	return strconv.Itoa(mine.Size())
}
