package testutils

// for testing access. see tests/files/access*.go

// NOTE: non-package variables cannot be overridden, except during init().
var (
	TestVar1 int
	testVar2 int
)

func init() {
	TestVar1 = 123
	testVar2 = 456
}

type TestAccessStruct struct {
	PublicField  string
	privateField string
}

func (tas TestAccessStruct) PublicMethod() string {
	return tas.PublicField + "/" + tas.privateField
}

func (tas TestAccessStruct) privateMethod() string {
	return tas.PublicField + "/" + tas.privateField
}

func NewTestAccessStruct(pub, priv string) TestAccessStruct {
	return TestAccessStruct{
		PublicField:  pub,
		privateField: priv,
	}
}

// see access6.g0 etc.
type PrivateInterface interface {
	privateMethod() string
}

func PrintPrivateInterface(pi PrivateInterface) {
	println("testutils.PrintPrivateInterface", pi.privateMethod())
}
