package main

type T interface {
	~int
}

func main() {}

// Error:
// parse_tilde_iface.gno:4:2: tilde operator is not permitted
