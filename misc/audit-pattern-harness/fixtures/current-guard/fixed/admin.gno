package admin

var owner = address("g1qz8e0fz3y0pl9y4dq9d7c5dwnyu6qf04hs7z0a")

func TransferOwnership(cur realm, next address) {
	transfer(0, cur, next)
}

func transfer(_ int, rlm realm, next address) {
	// Fixed: reject any realm value that is not the live acting realm.
	if !rlm.IsCurrent() {
		panic("invalid realm")
	}
	if rlm.Previous().Address() != owner {
		panic("owner only")
	}
	owner = next
}

func Owner() address {
	return owner
}
