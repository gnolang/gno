package gnolang

func (StringValue) IsMutable() bool   { return false }
func (BigintValue) IsMutable() bool   { return false }
func (BigdecValue) IsMutable() bool   { return false }
func (DataByteValue) IsMutable() bool { return true }
func (pv PointerValue) IsMutable() bool {
	if pv.Base == nil {
		return false
	}
	return pv.Base.IsMutable()
}
func (*ArrayValue) IsMutable() bool { return true }
func (sv *SliceValue) IsMutable() bool {
	_, ok := sv.Base.(ReadonlyValue)
	return !ok
}
func (*StructValue) IsMutable() bool      { return true }
func (*FuncValue) IsMutable() bool        { return false }
func (*MapValue) IsMutable() bool         { return true }
func (*BoundMethodValue) IsMutable() bool { return false }
func (TypeValue) IsMutable() bool         { return false }
func (*PackageValue) IsMutable() bool     { return false }
func (*Block) IsMutable() bool            { return true }
func (RefValue) IsMutable() bool {
	panic("should not happen")
}
func (*HeapItemValue) IsMutable() bool { return true }
func (ReadonlyValue) IsMutable() bool  { return false }
