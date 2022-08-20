package root

var (
	counter               int
	currentImplementation = "gno.land/r/tests/upgrade-c/v1"
)

func Inc() int {
	// TODO: if caller is currentImplementation
	counter++
	return counter
}

func Get() int {
	return counter
}

func UpdateCurrentImplementation(address string) {
	// TODO: if is admin
	currentImplementation = address
}
