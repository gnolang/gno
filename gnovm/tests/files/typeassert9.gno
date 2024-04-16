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
// *cannot use _ as a value or type*
