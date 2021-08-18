// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data-structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
	tree, _ = tree.Put(types.String("key2"), "value2")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 4

// Realm:
// c[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:11]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 11,
//   _OwnerNewTime: (uint64) 6,
//   _ModTime: (uint64) 0,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 33                                       |key3|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 33                                 |value3|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  01 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:6]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 6,
//   _OwnerNewTime: (uint64) 5,
//   _ModTime: (uint64) 7,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 32                                       |key2|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 32                                 |value2|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  02 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 4AF0F175D54357F0FEEAE4CF180A42BE848369E8,
//      ValueImage: (gno.RefImage) {
//       RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//       NewTime: (uint64) 11,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  6d 94 63 de 08 93 c2 bc  ea f1 f4 bd e2 89 b4 3c  |m.c............<|
//         00000010  d7 d2 33 34                                       |..34|
//        }
//       }
//      }
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:4]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 4,
//   _OwnerNewTime: (uint64) 5,
//   _ModTime: (uint64) 9,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 30                                       |key0|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 30                                 |value0|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  01 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:5]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 5,
//   _OwnerNewTime: (uint64) 0,
//   _ModTime: (uint64) 8,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 31                                       |key1|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 31                                 |value1|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  03 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 4AF0F175D54357F0FEEAE4CF180A42BE848369E8,
//      ValueImage: (gno.RefImage) {
//       RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//       NewTime: (uint64) 4,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  bd 34 c8 dc d7 ad fc 5a  5f 09 1e a9 34 e7 c0 2c  |.4.....Z_...4..,|
//         00000010  30 c9 4f bb                                       |0.O.|
//        }
//       }
//      }
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 4AF0F175D54357F0FEEAE4CF180A42BE848369E8,
//      ValueImage: (gno.RefImage) {
//       RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//       NewTime: (uint64) 6,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  58 3a d3 94 da 01 33 2f  b7 33 6d 4c 80 c8 67 57  |X:....3/.3mL..gW|
//         00000010  39 b3 62 fa                                       |9.b.|
//        }
//       }
//      }
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:4]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 4,
//   _OwnerNewTime: (uint64) 5,
//   _ModTime: (uint64) 9,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 30                                       |key0|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 30                                 |value0|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  01 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=(gno.BlockValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 0,
//   _OwnerNewTime: (uint64) 0,
//   _ModTime: (uint64) 10,
//   _RefCount: (int) 0
//  },
//  ParentID: (gno.ObjectID) OIDNONE:0,
//  Values: ([]gno.TypedValueImage) (len=3 cap=3) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//    ValueImage: (gno.FuncValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//     IsMethod: (bool) false,
//     Name: (gno.Name) (len=6) "init.0",
//     ClosureRef: (gno.RefImage) {
//      RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//      NewTime: (uint64) 0,
//      Hash: (gno.ValueHash) {
//       Hashlet: (gno.Hashlet) (len=20 cap=20) {
//        00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//        00000010  00 00 00 00                                       |....|
//       }
//      }
//     },
//     FileName: (gno.Name) (len=16) "files/zrealm7.go",
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
//      RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//      NewTime: (uint64) 0,
//      Hash: (gno.ValueHash) {
//       Hashlet: (gno.Hashlet) (len=20 cap=20) {
//        00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//        00000010  00 00 00 00                                       |....|
//       }
//      }
//     },
//     FileName: (gno.Name) (len=16) "files/zrealm7.go",
//     PkgPath: (string) (len=15) "gno.land/r/test"
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 4AF0F175D54357F0FEEAE4CF180A42BE848369E8,
//      ValueImage: (gno.RefImage) {
//       RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//       NewTime: (uint64) 5,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  cf 16 19 63 6e 34 5c 18  81 4a ee c9 7d 24 cb 4c  |...cn4\..J..}$.L|
//         00000010  0d 3f d0 fc                                       |.?..|
//        }
//       }
//      }
//     }
//    }
//   }
//  }
// }
//
// d[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:6]
// d[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:5]
