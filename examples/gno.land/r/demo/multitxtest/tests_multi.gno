package multitxtest

var slice = []string{"undead-element"}

func Pop() {
	slice = slice[:len(slice)-1]
}

func Push() {
	slice = append(slice, "new-element")
}

func GetSlice() []string {
	return slice
}
