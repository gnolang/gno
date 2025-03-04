// hello is a package for testing
package hello

// sl is an int array
var sl = []int{1, 2, 3, 4, 5}

func fn() func(string) string { return Echo }

// myStruct is a struct for testing
type myStruct struct{ a int }

var myStructInst = myStruct{a: 1000}

// Foo is a method for testing
func (ms myStruct) Foo() string { return "myStruct.Foo" }

// Panic is a func for testing
func Panic() { panic("foo") }

var counter int = 42

// Test public and private vars
var (
	pvString  = "private string" // A private var
	PubString = "public string"
)

const ConstString = "const string"

func Echo(msg string) (res string) { res = "echo:" + msg; return }
func GetCounter() int              { return counter }
func Inc() int                     { counter += 1; return counter }
func pvEcho(msg string) string     { return "pvecho:" + msg }
