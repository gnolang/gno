package administrable

import "testing"

func TestPackage(t *testing.T) {
	admin := New()
	println(admin.List())
}

func ExamplePackageLevel() {
}

func ExampleEmbedding() {
	println("test")
}
