// PKGPATH: gno.land/r/test
package test

var root *Node

type Key interface{}

type Node struct {
	Key   Key
	Left  *Node `gno:owned`
	Right *Node `gno:owned`
}

func init() {
	root = &Node{
		Key: "old",
	}
}

func main() {
	root = &Node{
		Key: "new",
	}
}

// Realm:
// c[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:4]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   ID: (gno.ObjectID) OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:4,
//   OwnerID: (gno.ObjectID) OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0,
//   ModTime: (uint64) 0,
//   RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=3 cap=3) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=3 cap=8) {
//     00000000  6e 65 77                                          |new|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//    ValueImage: (gno.ValueImage) <nil>
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//    ValueImage: (gno.ValueImage) <nil>
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=(gno.BlockValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   ID: (gno.ObjectID) OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0,
//   OwnerID: (gno.ObjectID) OIDNONE:0,
//   ModTime: (uint64) 3,
//   RefCount: (int) 0
//  },
//  ParentID: (gno.ObjectID) OIDNONE:0,
//  Values: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 1AF40977153D0FABAB9803BF33EDEBA8EB420CC5,
//    ValueImage: (gno.TypeValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) B06B716FF82D41A482D5C1CC3711002B74717639
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 1AF40977153D0FABAB9803BF33EDEBA8EB420CC5,
//    ValueImage: (gno.TypeValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 8F3FCA65F6CA73D096C06F68E24FF93EA462D350
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//    ValueImage: (gno.FuncValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//     IsMethod: (bool) false,
//     Name: (gno.Name) (len=6) "init.2",
//     ClosureRef: (gno.RefImage) {
//      _ID: (gno.ObjectID) OIDNONE:0,
//      Hash: (gno.ValueHash) {
//       Hashlet: (gno.Hashlet) (len=20 cap=20) {
//        00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//        00000010  00 00 00 00                                       |....|
//       }
//      }
//     },
//     FileName: (gno.Name) (len=16) "files/zrealm3.go",
//     PkgPath: (string) (len=15) "gno.land/r/test"
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//    ValueImage: (gno.FuncValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//     IsMethod: (bool) false,
//     Name: (gno.Name) (len=4) "main",
//     ClosureRef: (gno.RefImage) {
//      _ID: (gno.ObjectID) OIDNONE:0,
//      Hash: (gno.ValueHash) {
//       Hashlet: (gno.Hashlet) (len=20 cap=20) {
//        00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//        00000010  00 00 00 00                                       |....|
//       }
//      }
//     },
//     FileName: (gno.Name) (len=16) "files/zrealm3.go",
//     PkgPath: (string) (len=15) "gno.land/r/test"
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 7B2E21E5A17CE618ADA4860C549E3E24B9D73269,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 8F3FCA65F6CA73D096C06F68E24FF93EA462D350,
//      ValueImage: (gno.RefImage) {
//       _ID: (gno.ObjectID) OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:4,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  78 17 a3 62 53 c7 38 55  e0 a7 06 6d 97 75 25 12  |x..bS.8U...m.u%.|
//         00000010  68 2d b1 d4                                       |h-..|
//        }
//       }
//      }
//     }
//    }
//   }
//  }
// }
//
// d[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:2]
