package t56a

var count int

func Count() int { return count }

// Helper writes through a local helper: its call is marked at preprocess.
func Helper(cur realm) int {
	f := func() { count++ }
	f()
	return count
}

// NotHelper assigns f twice, so f is no helper and its call is refused.
func NotHelper(cur realm) (refused bool, n int) {
	func() {
		defer func() { refused = recover() != nil }()
		var f func()
		f = func() { count++ }
		f = func() { count += 100 }
		f()
	}()
	return refused, count
}
