package benchops

import "fmt"

// CPUOp represents a VM opcode for benchmarking.
// Values mirror gnolang.Op but are defined here to avoid circular imports.
type CPUOp byte

const CPUOpInvalid CPUOp = 0x00

// StoreOp represents a store operation for benchmarking.
type StoreOp byte

// store code
const (
	StoreOpInvalid StoreOp = 0x00 // invalid

	// gno store
	StoreGetObject       StoreOp = 0x01 // get value and unmarshl to object from store
	StoreSetObject       StoreOp = 0x02 // marshal object and set value in store
	StoreDeleteObject    StoreOp = 0x03 // delete value from store
	StoreGetPackage      StoreOp = 0x04 // get package from store
	StoreSetPackage      StoreOp = 0x05 // get package from store
	StoreGetType         StoreOp = 0x06 // get type from store
	StoreSetType         StoreOp = 0x07 // set type in store
	StoreGetBlockNode    StoreOp = 0x08 // get block node from store
	StoreSetBlockNode    StoreOp = 0x09 // set block node in store
	StoreAddMemPackage   StoreOp = 0x0A // add mempackage to store
	StoreGetMemPackage   StoreOp = 0x0B // get mempackage from store
	StoreGetPackageRealm StoreOp = 0x0C // add mempackage to store
	StoreSetPackageRealm StoreOp = 0x0D // get mempackage from store

	AminoMarshal    StoreOp = 0x0E // marshal mem package and realm to binary
	AminoMarshalAny StoreOp = 0x0F // marshal gno object to binary
	AminoUnmarshal  StoreOp = 0x10 // unmarshl binary to gno object, package and realm

	// underlying store
	StoreGet StoreOp = 0x11 // Get binary value by key
	StoreSet StoreOp = 0x12 // Set binary value by key

	FinalizeTx StoreOp = 0x13 // finalize transaction

	// realm operations
	RealmDidUpdate  StoreOp = 0x14 // realm dirty tracking / escape analysis
	RealmFinalizeTx StoreOp = 0x15 // realm transaction finalization

	invalidStoreCode string = "StoreInvalid"
)

// NativeOp represents a native operation for benchmarking.
type NativeOp byte

// native code
const (
	NativeOpInvalid NativeOp = 0x00 // invalid

	NativePrint       NativeOp = 0x01 // print to console
	NativePrint_1     NativeOp = 0x02 // print to console
	NativePrint_1000  NativeOp = 0x03
	NativePrint_10000 NativeOp = 0x04 // print 1000 times to console

	invalidNativeCode string = "NativeInvalid"
)

func GetNativePrintCode(size int) NativeOp {
	switch size {
	case 1:
		return NativePrint_1
	case 1000:
		return NativePrint_1000
	case 10000:
		return NativePrint_10000
	default:
		panic(fmt.Sprintf("invalid print size: %d", size))
	}
}

// the index of the code string should match with the constant code number above.
var storeCodeNames = []string{
	invalidStoreCode,
	"StoreGetObject",
	"StoreSetObject",
	"StoreDeleteObject",
	"StoreGetPackage",
	"StoreSetPackage",
	"StoreGetType",
	"StoreSetType",
	"StoreGetBlockNode",
	"StoreSetBlockNode",
	"StoreAddMemPackage",
	"StoreGetMemPackage",
	"StoreGetPackageRealm",
	"StoreSetPackageRealm",
	"AminoMarshal",
	"AminoMarshalAny",
	"AminoUnMarshal",
	"StoreGet",
	"StoreSet",
	"FinalizeTx",
	"RealmDidUpdate",
	"RealmFinalizeTx",
}

var nativeCodeNames = []string{
	invalidNativeCode,
	"NativePrint",
	"NativePrint_1",
	"NativePrint_1000",
	"NativePrint_10000",
}

type Code [2]byte
type Type byte

const (
	TypeOpCode Type = 0x01
	TypeStore  Type = 0x02
	TypeNative Type = 0x03
)

func VMOpCode(opCode CPUOp) Code {
	return [2]byte{byte(TypeOpCode), byte(opCode)}
}

func StoreCode(storeCode StoreOp) Code {
	return [2]byte{byte(TypeStore), byte(storeCode)}
}

func NativeCode(nativeCode NativeOp) Code {
	return [2]byte{byte(TypeNative), byte(nativeCode)}
}

func StoreCodeString(storeCode StoreOp) string {
	if int(storeCode) >= len(storeCodeNames) {
		return invalidStoreCode
	}
	return storeCodeNames[storeCode]
}

func NativeCodeString(nativeCode NativeOp) string {
	if int(nativeCode) >= len(nativeCodeNames) {
		return invalidNativeCode
	}
	return nativeCodeNames[nativeCode]
}
