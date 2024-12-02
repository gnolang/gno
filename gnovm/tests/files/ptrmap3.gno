// PKGPATH: gno.land/r/ptr_map
package ptr_map

var mapus map[uint64]string = make(map[uint64]string)

func init() {
	mapus[3] = "three"
	mapus[5] = "five"
	mapus[9] = "nine"
}

func FindMapWithKey(k uint64) string {
	return mapus[k]
}

func main() {
	r := FindMapWithKey(3)
	println(r)
}

// Output:
// three
