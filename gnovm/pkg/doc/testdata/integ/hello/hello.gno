// hello is a package for testing
package hello

// sl is an int array
var sl = []int{1, 2, 3, 4, 5}

func fn() func(string) string { return Echo }

// myStruct is a struct for testing
type myStruct struct {
	// a is a field
	a int // a comment
}

var myStructInst = myStruct{a: 1000}

// This var belongs to the myStruct type
var myStructPtr *myStruct

// This const belongs to the myStruct type
const myStructConst *myStruct = &myStruct{a: 1000}

// Foo is a method for testing
func (ms myStruct) Foo() string { return "myStruct.Foo" }

// Test type aliases
type myAlias = myStruct

// Test array type
type myArrayType [5]int

// Test slice type
type mySliceType []int

// Test map type
type myMapType map[string]int

// Test chan type
type myChanType chan int

// Test func type
type myFuncType func(int) string

// Test pointer type
type myPointerType *myStruct

// Panic is a func for testing
func Panic() { panic("foo") }

// BUG(who): Bug about myInterface

// myInterface is an interface for testing
type myInterface interface {
	error
	// Bar is for testing
	Bar(x int) string // Bar line comment
}

var counter int = 42

// Test public and private vars
var (
	pvString  = "private string" // A private var
	PubString = "public string"
)

const ConstString = "const string"

func Echo(msg string) (res string) { res = "echo:" + msg; return }
func GetCounter() int              { return counter }
func Inc(cur realm) int            { counter += 1; return counter }
func pvEcho(msg string) string     { return "pvecho:" + msg }
