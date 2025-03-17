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
	defer func() {
		if r := recover(); r != nil {
			println(r)
		} else {
			println("ok")
		}
	}()

	_ = i.(Setter)
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
// interface conversion: interface is nil, not main.Setter
// interface conversion: interface is nil, not main.Setter
// ok
// main.ValueSetter doesn't implement interface {Set func(string)} (method Set has pointer receiver)
// ok
