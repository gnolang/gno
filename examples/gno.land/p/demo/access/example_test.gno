package access

func ExamplePackageLevel() {
	// set this globally or during init()
	acl := New()

	// in functions
	acl.AssertCurrentHasAccess()
}

func ExampleEmbedding() {
	// declare a new struct, and embed access.Set.
	type MyObject struct {
		myField int
		Set
	}

	// initialize the object, it now has its own admin set.
	myObject := MyObject{
		myField: 42,
		Set:     New(),
	}

	// check from the context of the object.
	myObject.Set.AssertCurrentHasAccess()
}
