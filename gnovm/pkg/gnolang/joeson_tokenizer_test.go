package gnolang

import (
	"fmt"
	"testing"
)

func TestTokenize1(t *testing.T) {
	source := `
package p
import fmt "fmt"
const pi = 3.14
type T struct{
	a int
	 b string
	  c float
}
var x int
func f() { L: }

var (
	_ int = 23
	_ string = "abc"
)

type (
	Alpha struct {
		a int
		b string
		 c float
	}
	Beta struct {
		a []string{256} // comment
	}
	 Gamma struct {}
)

func f() {
	if true {
		if false {
			// after 1 below, there should have an automatic ; inserted
			n := 1
			fmt.Println("no")
		}
	}
	a := []string{
		"foo",
		 "bar",
	 "baz",
	}
}
	`
	fmt.Println(source)
	if code, e := tokenize(source); e != nil {
		t.Error(e.Error())
	} else {
		work := code.PrintWorkText()
		fmt.Println(work)

		// no test yet during development phase
		// it's just printing what we obtain.
		// if work != expectWork {
		// 	t.Fail()
		// }
	}
}

const expectWork string = `package p;
import fmt "fmt";
const pi= 3.14;
type T struct{a int;
b string;
c float;
};
var x int;
func f(){L:};
var(_ int= 23;
_ string= "abc";
);
type(Alpha struct{a int;
b string;
c float;
};
Beta struct{a[]string{256};
};
Gamma struct{};
);
func f(){if true{if false{n:= 1;
fmt.Println("no");
};
};
a:=[]string{"foo","bar","baz",};
};
`
