package crossrealm

var m map[string]int

func init() {
	m = make(map[string]int)
	m["a"] = 1
}

func GetMap() map[string]int {
	return m
}
