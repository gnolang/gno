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

func cmpSetter(i interface{}) {
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
		i              interface{}
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
// non-concrete interface{} doesn't implement interface{Set func(string)()}
// non-concrete main.Setter doesn't implement interface{Set func(string)()}
// ok
// main.ValueSetter doesn't implement interface{Set func(string)()}
// ok
