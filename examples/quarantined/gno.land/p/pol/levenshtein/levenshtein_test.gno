package levenshtein

import "testing"

func TestExample(t *testing.T) {
	if Distance("float", "boats") != 3 {
		t.Fatal("expected float vs boats to be 3")
	}

	if Distance("kitten", "sitting") != 3 {
		t.Fatal("expected kitten vs sitting to be 3")
	}

	if Distance("foo", "foo") != 0 {
		t.Fatal("expected foo and foo to be 0 (the same)")
	}
}

type person struct {
	name     string
	lastName string
}

func (p person) LString() string {
	return p.name + " " + p.lastName
}

func TestWithMyPhoneContact(t *testing.T) {
	contacts := []Levenshteinable{
		person{"maman", ""},
		person{"papa", ""},
		person{"paul", "parisot"},
		person{"leonardo", "dicaprio"},
		person{"jenifer", "lawrence"},
		person{"jae", "kwon"},
		person{"justin", "bieber"},
	}
	who, _ := PickFromString("pau parosit", contacts)
	// println(who.LString(), n)
	if who.LString() != "paul parisot" {
		t.Fatal("expected to be paul parisot !")
	}

	who, _ = PickFromString("jkwo0n", contacts)
	// println(who.LString(), n)
	if who.LString() != "jae kwon" {
		t.Fatal("expected to be jae kwon !")
	}

	who, _ = PickFromString("paulparis", contacts)
	// println(who, n)
}

func TestWithAVL(t *testing.T) {
	contacts := []Levenshteinable{
		person{"maman", ""},
		person{"papa", ""},
		person{"paul", "parisot"},
		person{"leonardo", "dicaprio"},
		person{"jenifer", "lawrence"},
		person{"jae", "kwon"},
		person{"justin", "bieber"},
	}
	list := SeveralPickFromString("paulparis", contacts, 3)
	if len(list) != 3 {
		t.Fatal("expected 3 elements, got ", len(list))
	}
	if list[0] != contacts[2] {
		t.Fatal("expected first element to be paul parisot")
	}
	if list[1] != contacts[1] {
		t.Fatal("expected second to be papa")
	}
}

func TestWhitespaces(t *testing.T) {
	if Distance("hello world", "helloworld") != 1 {
		t.Fatal("'helloworld' vs 'hello world' doesnt have 1 distance")
	}
	if Distance("hello  world", "hello world") != 1 {
		t.Fatal("'hello  world' vs 'hello world' doesnt have 1 distance")
	}
	if Distance("hello world", "hello  world") != 1 {
		t.Fatal("'hello world' vs 'hello  world' doesnt have 1 distance")
	}
}
