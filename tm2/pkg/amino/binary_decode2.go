package amino

// DecodeFieldNumberAndTyp3 reads a protobuf field key (field number + wire type).
// This is the exported wrapper around decodeFieldNumberAndTyp3 for use by
// generated code (genproto2).
func DecodeFieldNumberAndTyp3(bz []byte) (num uint32, typ Typ3, n int, err error) {
	return decodeFieldNumberAndTyp3(bz)
}
