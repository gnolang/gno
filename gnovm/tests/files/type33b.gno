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
// nil is not of type main.ex
