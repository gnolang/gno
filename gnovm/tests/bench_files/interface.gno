package main

// Tests interface dynamic dispatch vs direct calls.
// Expected: Dynamic dispatch may cost more than direct calls.

type Speaker interface {
	Speak() string
}

type Dog struct{ name string }

func (d Dog) Speak() string { return "woof" }

type Cat struct{ name string }

func (c Cat) Speak() string { return "meow" }

func main() {
	DirectCall()
	InterfaceCall()
	TypeAssertion()
}

func DirectCall() {
	d := Dog{name: "Rex"}
	_ = d.Speak()
	c := Cat{name: "Whiskers"}
	_ = c.Speak()
}

func InterfaceCall() {
	var s Speaker
	s = Dog{name: "Rex"}
	_ = s.Speak()
	s = Cat{name: "Whiskers"}
	_ = s.Speak()
}

func TypeAssertion() {
	var s Speaker = Dog{name: "Rex"}
	if d, ok := s.(Dog); ok {
		_ = d.name
	}
}

// Gasprofile:
// Opcodes:
//   OpAssign: count=7 gas=553
//   OpBody: count=24 gas=1032
//   OpCall: count=8 gas=2048
//   OpCompositeLit: count=5 gas=250
//   OpDefine: count=3 gas=333
//   OpEval: count=110 gas=3190
//   OpExec: count=11 gas=275
//   OpFieldType: count=9 gas=531
//   OpFuncType: count=8 gas=648
//   OpHalt: count=76 gas=76
//   OpIfCond: count=1 gas=38
//   OpInterfaceType: count=1 gas=75
//   OpPopBlock: count=68 gas=204
//   OpPopResults: count=4 gas=4
//   OpPrecall: count=8 gas=1656
//   OpReturn: count=4 gas=152
//   OpReturnFromBlock: count=4 gas=144
//   OpSelector: count=5 gas=160
//   OpStructLit: count=5 gas=895
//   OpStructType: count=2 gas=348
//   OpTypeAssert2: count=1 gas=34
//   OpTypeDecl: count=3 gas=300
//   OpValueDecl: count=2 gas=226
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   interface.gno:6 -: count=3 gas=154
//   interface.gno:10 -: count=3 gas=154
//   interface.gno:14 -: count=3 gas=154
//   interface.gno:25 DirectCall: count=6 gas=427
//   interface.gno:26 DirectCall: count=7 gas=661
//   interface.gno:27 DirectCall: count=6 gas=427
//   interface.gno:28 DirectCall: count=7 gas=661
//   interface.gno:32 InterfaceCall: count=3 gas=167
//   interface.gno:33 InterfaceCall: count=6 gas=395
//   interface.gno:34 InterfaceCall: count=7 gas=661
//   interface.gno:35 InterfaceCall: count=6 gas=395
//   interface.gno:36 InterfaceCall: count=7 gas=661
//   interface.gno:40 TypeAssertion: count=8 gas=483
//   interface.gno:41 TypeAssertion: count=8 gas=324
//   interface.gno:42 TypeAssertion: count=4 gas=169
// SubOps:
//   AssignVar: count=7
//   DefineVar: count=4
//   ParamAssign: count=4
//   StructField: count=5
// Variables:
//   interface.gno:12 #0: count=2
//   interface.gno:16 #0: count=2
//   interface.gno:25 d: count=1
//   interface.gno:25 name: count=1
//   interface.gno:26 _: count=1
//   interface.gno:27 c: count=1
//   interface.gno:27 name: count=1
//   interface.gno:28 _: count=1
//   interface.gno:33 name: count=1
//   interface.gno:33 s: count=1
//   interface.gno:34 _: count=1
//   interface.gno:35 name: count=1
//   interface.gno:35 s: count=1
//   interface.gno:36 _: count=1
//   interface.gno:40 name: count=1
//   interface.gno:41 d: count=1
//   interface.gno:41 ok: count=1
//   interface.gno:42 _: count=1
