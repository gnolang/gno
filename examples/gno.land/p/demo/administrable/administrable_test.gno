package administrable

import "testing"

func TestPackage(t *testing.T){
	var admin = New()
	println(admin.List())
}
