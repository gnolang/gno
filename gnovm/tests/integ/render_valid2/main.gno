package render_valid2

type Test struct{}

// 'Render' method does not need to follow the signature of the pkg scope 'Render' function

func (t *Test) Render(input int) int {
	return input
}
