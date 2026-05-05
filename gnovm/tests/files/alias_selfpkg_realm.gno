// PKGPATH: gno.land/r/aliasself
package aliasself

// self-declared DeclaredType and an alias to it within the same package.
type T struct {
	Name string
}

type MyAlias = T

func main(cur realm) {
	var t MyAlias = MyAlias{Name: "hi"}
	println(t.Name)
}

// Output:
// hi

// Types:
// T[gno.land/r/aliasself.T]=
//     {
//         "ID": "gno.land/r/aliasself.T"
//     }
// MyAlias[gno.land/r/aliasself.T]=
//     {
//         "ID": "gno.land/r/aliasself.T"
//     }
