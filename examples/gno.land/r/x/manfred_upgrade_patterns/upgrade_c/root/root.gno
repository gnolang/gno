package root

var (
	counter               int
	currentImplementation = "gno.land/r/x/manfred_upgrade_patterns/upgrade_c/v1"
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
