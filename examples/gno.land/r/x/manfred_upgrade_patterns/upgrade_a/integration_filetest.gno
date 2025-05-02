package main

import (
	v1 "gno.land/r/x/manfred_upgrade_patterns/upgrade_a/v1"
	v2 "gno.land/r/x/manfred_upgrade_patterns/upgrade_a/v2"
)

func main() {
	println(v1.GetString())
	cross(v1.Inc)(10)
	cross(v1.Inc)(100)
	println(v1.GetString())

	println(v2.GetString())
	cross(v2.Inc)(1000)
	cross(v2.Inc)(10000)
	println(v2.GetString())

	println(v1.GetString())
	cross(v1.Inc)(10)
	cross(v1.Inc)(100)
	println(v1.GetString())

	println(v2.GetString())
	cross(v2.Inc)(1000)
	cross(v2.Inc)(10000)
	println(v2.GetString())

	println("done")
}

// Output:
// v1 - 0
// v1 - 110
// v2 - 110
// v2 - 11110
// v1 - 110
// v1 - 220
// v2 - 11220
// v2 - 22220
// done
