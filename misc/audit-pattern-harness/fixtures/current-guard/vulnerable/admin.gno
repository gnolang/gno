package admin

var owner = address("g1qz8e0fz3y0pl9y4dq9d7c5dwnyu6qf04hs7z0a")

func TransferOwnership(cur realm, next address) {
	// Vulnerable: reads cur.Previous() without first checking cur.IsCurrent().
	if cur.Previous().Address() != owner {
		panic("owner only")
	}
	owner = next
}

func Owner() address {
	return owner
}
