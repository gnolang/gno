package main

type Outfit struct {
	Scarf string
	Shirt string
	Belts string
	Strap string
	Pants string
	Socks string
	Shoes string
}

func main() {
	s := Outfit{
		// some fields are out of order.
		// some fields are left unset.
		Scarf: "scarf",
		Shirt: "shirt",
		Shoes: "shoes",
		Socks: "socks",
	}
	// some fields out of order are used.
	// some fields left unset are used.
	print(s.Shoes + "," + s.Shirt + "," + s.Pants + "," + s.Scarf)
}

// Output:
// shoes,shirt,,scarf
