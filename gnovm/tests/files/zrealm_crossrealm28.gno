// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

type Struct struct {
	A int
}

var s *Struct

func init() {
	s = &Struct{A: 100}
	crossrealm_b.SetObject(cross, func() { println(&s.A) })
}

func main(cur realm) {
	//s := &Struct{A: 100}
	crossrealm_b.SetObject(cross, 123)
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm_test.Struct)

// Preprocessed:
// file{ package crossrealm_test; import crossrealm_b gno.land/r/demo/tests/crossrealm_b; type Struct (const-type gno.land/r/crossrealm_test.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm_test.Struct}); func init.2() { s<~VPBlock(3,1)> = &((const-type gno.land/r/crossrealm_test.Struct){A<VPField(0,0,A)>: (const (100 int))}); (const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject((const (undefined)), func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm_test) package{})).s.A)) }) }; func main(cur (const-type .uverse.realm)) { (const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject((const (undefined)), (const (123 int))); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

// Realm:
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:13]=
//     @@ -1,20 +1,15 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:13",
//     -        "ModTime": "15",
//     +        "ModTime": "16",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:2",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "ewAAAAAAAAA=",
//              "T": {
//     -            "@type": "/gno.FuncType",
//     -            "Params": null,
//     -            "Results": null
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "47596878835b82fc5bcd8469e6d3f9b90e545e45",
//     -            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:16"
//     +            "@type": "/gno.PrimitiveType",
//     +            "value": "32"
//              }
//          }
//      }
// d[0edc46caf30c00efd87b6c272673239eafbd051e:16]
// finalizerealm["gno.land/r/crossrealm_test"]
// u[f5a516808f8976c33939133293d598ce3bca4e8d:8]=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "ZAAAAAAAAAA=",
//     +            "N": "ewAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -10,7 +10,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:8",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//              "RefCount": "1"
//          }
