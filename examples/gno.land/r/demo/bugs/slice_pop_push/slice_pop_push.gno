package slice_pop_push

var slice = []string{"undead-element"}

func Pop() {
	slice = slice[:len(slice)-1]
}

func Push() {
	slice = append(slice, "new-element")
}

func Render(path string) string {
	os := []string{"undead-element-2"}
	os = os[:len(os)-1]
	os = append(os, "new-element-2")

	return os[0] + slice[0]
}
