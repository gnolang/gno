package crossrealm

var S []Fooer

func SetSlice(fs []Fooer) {
	S = fs
	for _, f := range fs {
		println(f)
	}
}
