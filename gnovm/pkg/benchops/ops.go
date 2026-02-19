package benchops

import "fmt"

// store code
const (
	// gno store
	StoreGetObject       byte = 0x01 // get value and unmarshl to object from store
	StoreSetObject       byte = 0x02 // marshal object and set value in store
	StoreDeleteObject    byte = 0x03 // delete value from store
	StoreGetPackage      byte = 0x04 // get package from store
	StoreSetPackage      byte = 0x05 // get package from store
	StoreGetType         byte = 0x06 // get type from store
	StoreSetType         byte = 0x07 // set type in store
	StoreDeleteType      byte = 0x08 // delete type from store
	StoreGetBlockNode    byte = 0x09 // get block node from store
	StoreSetBlockNode    byte = 0x0A // set block node in store
	StoreAddMemPackage   byte = 0x0B // add mempackage to store
	StoreGetMemPackage   byte = 0x0C // get mempackage from store
	StoreGetPackageRealm byte = 0x0D // add mempackage to store
	StoreSetPackageRealm byte = 0x0E // get mempackage from store
	AminoMarshal         byte = 0x0F // marshal mem package and realm to binary
	AminoMarshalAny      byte = 0x10 // marshal gno object to binary
	AminoUnmarshal       byte = 0x11 // unmarshl binary to gno object, package and realm

	// underlying store
	StoreGet byte = 0x12 // Get binary value by key
	StoreSet byte = 0x13 // Set binary value by key

	FinalizeTx byte = 0x14 // finalize transaction

	invalidStoreCode string = "StoreInvalid"
)

// native code
const (
	NativePrint       byte = 0x01 // print to console
	NativePrint_1     byte = 0x02 // print to console
	NativePrint_1000  byte = 0x03
	NativePrint_10000 byte = 0x04 // print 1000 times to console

	invalidNativeCode string = "NativeInvalid"
)

func GetNativePrintCode(size int) byte {
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

func VMOpCode(opCode byte) Code {
	return [2]byte{byte(TypeOpCode), opCode}
}

func StoreCode(storeCode byte) Code {
	return [2]byte{byte(TypeStore), storeCode}
}

func NativeCode(nativeCode byte) Code {
	return [2]byte{byte(TypeNative), nativeCode}
}

func StoreCodeString(storeCode byte) string {
	if int(storeCode) >= len(storeCodeNames) {
		return invalidStoreCode
	}
	return storeCodeNames[storeCode]
}

func NativeCodeString(nativeCode byte) string {
	if int(nativeCode) >= len(nativeCodeNames) {
		return invalidNativeCode
	}
	return nativeCodeNames[nativeCode]
}
