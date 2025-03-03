package main

type Setter interface {
	Set(string)
}

type SetterClone interface {
	Set(string)
}

type ValueSetter struct {
	value string
}

func (s *ValueSetter) Set(value string) {
	s.value = value
}

func cmpSetter(i any) {
	if _, ok := i.(Setter); ok {
		println("ok")
	} else {
		println("not ok")
	}
}

func main() {
	var (
		i              any
		setter         Setter
		setterClone    SetterClone
		valueSetter    ValueSetter
		valueSetterPtr *ValueSetter
	)

	cmpSetter(i)

	i = setter
	cmpSetter(i)

	setterClone = valueSetterPtr
	setter = setterClone
	i = setter
	cmpSetter(i)

	i = valueSetter
	cmpSetter(i)

	i = valueSetterPtr
	cmpSetter(i)
}

// Output:
// not ok
// not ok
// ok
// not ok
// ok
