package benchmarking

// store code
const (
	StoreGetObject     byte = 0x01 // get value and unmarshl to object from store
	StoreSetObject     byte = 0x02 // marshal object and set value in store
	StoreDeleteObject  byte = 0x03 // delete value from store
	StoreGetPackage    byte = 0x04 // get package from store
	StoreSetPackage    byte = 0x05 // get package from store
	StoreGetType       byte = 0x06 // get type from store
	StoreSetType       byte = 0x07 // set type in store
	StoreGetBlockNode  byte = 0x08 // get block node from store
	StoreSetBlockNode  byte = 0x09 // set block node in store
	StoreAddMemPackage byte = 0x0A // add mempackage to store
	StoreGetMemPackage byte = 0x0B // get mempackage from store
	FinalizeTx         byte = 0x0C // finalize realm transaction

	AminoMarshal   byte = 0x0D // marshal go object to binary value
	AminoUnMarshal byte = 0x0E // unmarshl binary value to go object

	StoreGet byte = 0x0F // Get binary value by key
	StoreSet byte = 0x10 // Set binary value by key

	invalidStoreCode string = "StoreInvalid"
)

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
	"FinalizeTx",
	"AminoMarshal",
	"AminoUnMarshal",
	"StoreGet",
	"StoreSet",
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
