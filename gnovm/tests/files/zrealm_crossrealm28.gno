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
	cross(crossrealm_b.SetObject)(func() { println(&s.A) })
}

func main() {
	crossing()

	//s := &Struct{A: 100}
	cross(crossrealm_b.SetObject)(123)
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm_test.Struct)

// Preprocessed:
// file{ package crossrealm_test; import crossrealm_b gno.land/r/demo/tests/crossrealm_b; type Struct (const-type gno.land/r/crossrealm_test.Struct); var s<!~VPBlock(2,1)> *(Struct<VPBlock(2,0)>); func init.2() { s<~VPBlock(3,1)> = &(Struct<VPBlock(3,0)>{A<VPField(0,0,A)>: (const (100 int))}); (const (cross func(func(interface {})) func(interface {})))((const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject)(func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm_test) package{})).s.A)) }) }; func main() { (const (crossing func()))(); (const (cross func(func(interface {})) func(interface {})))((const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject)((const (123 int))); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

// Realm:
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:13](-98)=
//     @@ -1,22 +1,16 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:13",
//     -        "ModTime": "15",
//     -        "ObjectSize": "191",
//     +        "ModTime": "16",
//     +        "ObjectSize": "320",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:2",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "ewAAAAAAAAA=",
//              "T": {
//     -            "@type": "/gno.FuncType",
//     -            "IsClosure": false,
//     -            "Params": null,
//     -            "Results": null
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "5b8e03f7f8b6fe2399fb8f4a27dd5a1a49518df8",
//     -            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:16"
//     +            "@type": "/gno.PrimitiveType",
//     +            "value": "32"
//              }
//          }
//      }
// d[0edc46caf30c00efd87b6c272673239eafbd051e:16](-304)
// finalizerealm["gno.land/r/crossrealm_test"]
// u[f5a516808f8976c33939133293d598ce3bca4e8d:8](5)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "ZAAAAAAAAAA=",
//     +            "N": "ewAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -10,8 +10,8 @@
//          ],
//          "ObjectInfo": {
//              "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:8",
//     -        "ModTime": "0",
//     -        "ObjectSize": "0",
//     +        "ModTime": "8",
//     +        "ObjectSize": "214",
//              "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//              "RefCount": "1"
//          }
