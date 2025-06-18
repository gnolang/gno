package main

type ex int

func (ex) Error() string { return "" }

func main() {
	defer func() {
		r := _.(ex)
		println(r)
	}()
}

// Error:
// main/typeassert3.gno:9:8-9: cannot use _ as value or type

// TypeCheckError:
// main/typeassert3.gno:9:8: cannot use _ as value or type
