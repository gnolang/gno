package mapdelete

var mapus map[uint64]string = make(map[uint64]string)

func init() {
	mapus[3] = "three"
	mapus[5] = "five"
	mapus[9] = "nine"
}

func DeleteMap(_ realm, k uint64) {
	delete(mapus, k)
}

func GetMap(k uint64) bool {
	_, exist := mapus[k]
	return exist
}
