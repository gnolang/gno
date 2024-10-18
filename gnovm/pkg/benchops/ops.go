package benchmarking

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
	StoreGetBlockNode    byte = 0x08 // get block node from store
	StoreSetBlockNode    byte = 0x09 // set block node in store
	StoreAddMemPackage   byte = 0x0A // add mempackage to store
	StoreGetMemPackage   byte = 0x0B // get mempackage from store
	StoreGetPackageRealm byte = 0x0C // add mempackage to store
	StoreSetPackageRealm byte = 0x0D // get mempackage from store

	AminoMarshal    byte = 0x0E // marshal mem package and realm to binary
	AminoMarshalAny byte = 0x0F // marshal gno object to binary
	AminoUnmarshal  byte = 0x10 // unmarshl binary to gno object, package and realm

	// underlying store
	StoreGet byte = 0x11 // Get binary value by key
	StoreSet byte = 0x12 // Set binary value by key

	FinalizeTx byte = 0x13 // finalize transaction

	invalidStoreCode string = "StoreInvalid"
)

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

type Code [2]byte

func VMOpCode(opCode byte) Code {
	return [2]byte{opCode, 0x00}
}

func StoreCode(storeCode byte) Code {
	return [2]byte{0x00, storeCode}
}

func StoreCodeString(storeCode byte) string {
	if int(storeCode) >= len(storeCodeNames) {
		return invalidStoreCode
	}
	return storeCodeNames[storeCode]
}
