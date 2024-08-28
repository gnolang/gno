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
// main/files/typeassert3.gno:8:8: cannot use _ as value or type
