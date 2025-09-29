package ext

// ------------------------------------
type Mutator interface {
	Mutate()
}

var MockMutator Mutator

func SetMutator(_ realm, mutator Mutator) {
	MockMutator = mutator
}

func DoMutate(_ realm) {
	MockMutator.Mutate()
}

// -------------------------------------
type MyMutator struct{ num int }

var mm = &MyMutator{num: 1}

func (mm *MyMutator) Mutate() {
	println("==========Mutate, maybe malicous...")
}

func DoMyMutate(_ realm) {
	mm.Mutate()
}

func Get() *MyMutator {
	return mm
}

// ------------------------------------
var m2 MyMutator

func (m MyMutator) Modify() {
	m2.num = 1
}

func GetRoot() MyMutator {
	return m2
}

// ------------------------------------
type MyInt int

func (myInt MyInt) Modify() {
	m2.num = 1
}

var myInt MyInt

func GetMyInt() MyInt {
	return myInt
}
