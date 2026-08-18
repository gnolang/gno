package run_realm

var counter int

func main(cur realm) {
	counter++
	println("realm main called, counter:", counter)
}

func Increment(cur realm) {
	counter++
	println("counter:", counter)
}
