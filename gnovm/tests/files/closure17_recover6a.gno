package main

import "errors"

//func p(s interface{}) {
//	fmt.Printf("%T \n", s)
//	if v, ok := s.(string); ok {
//		println(v)
//		panic(v)
//	} else {
//		println("---")
//	}
//}

type error interface {
	Error() string
}

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func main() {
	//panic(New("wtf"))
	panic(errors.New("wtf"))
}

// Error:
// wtf
