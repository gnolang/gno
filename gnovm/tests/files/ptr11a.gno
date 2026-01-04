package main

type S struct{}

func (s S) String() string { return "hey" }

type Stringer interface {
	String() string
}

func main() {
	defer func() {
		rec := recover()
		if rec != nil {
			println("recovered", rec)
		}
	}()
	v := (*S)(nil)
	println(Stringer(v).String())
}

// Output:
// recovered nil pointer dereference
