package main

type ex int

func (ex) Error() string { return "" }

func main() {
	defer func() {
		e := recover()
		_ = e.(ex)
	}()
}

// Error:
// *cannot use _ as a value or type*
